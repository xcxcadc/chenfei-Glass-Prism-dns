package main

import (
	"sort"
	"strings"
)

func conflictingServiceIDs(serviceID string, routes map[string]string, services []Service) []string {
	selected := make(map[string]Service)
	for _, service := range services {
		if routes[service.ID] != "" {
			selected[service.ID] = service
		}
	}
	if _, exists := selected[serviceID]; !exists {
		return nil
	}
	result := []string{serviceID}
	seen := map[string]struct{}{serviceID: {}}
	for index := 0; index < len(result); index++ {
		left := selected[result[index]]
		leftDomains := routingDomains(left.Domains)
		for candidateID, candidate := range selected {
			if _, exists := seen[candidateID]; exists {
				continue
			}
			if domainsOverlap(leftDomains, routingDomains(candidate.Domains)) {
				seen[candidateID] = struct{}{}
				result = append(result, candidateID)
			}
		}
	}
	sort.Strings(result)
	return result
}

func domainsOverlap(left, right []string) bool {
	for _, leftDomain := range left {
		for _, rightDomain := range right {
			if leftDomain == rightDomain ||
				strings.HasSuffix(leftDomain, "."+rightDomain) ||
				strings.HasSuffix(rightDomain, "."+leftDomain) {
				return true
			}
		}
	}
	return false
}

func normalizeConflictingRoutes(previous, current map[string]string, services []Service) (map[string]string, bool) {
	normalized, changed := canonicalizeServiceRoutes(current, services)
	if len(normalized) < 2 {
		return normalized, changed
	}

	serviceByID := make(map[string]Service, len(services))
	for _, service := range services {
		serviceByID[service.ID] = service
	}
	serviceIDs := make([]string, 0, len(normalized))
	for serviceID, proxyID := range normalized {
		if proxyID != "" {
			if _, exists := serviceByID[serviceID]; exists {
				serviceIDs = append(serviceIDs, serviceID)
			}
		}
	}
	sort.Strings(serviceIDs)

	parent := make(map[string]string, len(serviceIDs))
	for _, serviceID := range serviceIDs {
		parent[serviceID] = serviceID
	}
	var find func(string) string
	find = func(serviceID string) string {
		if parent[serviceID] != serviceID {
			parent[serviceID] = find(parent[serviceID])
		}
		return parent[serviceID]
	}
	union := func(left, right string) {
		leftRoot := find(left)
		rightRoot := find(right)
		if leftRoot == rightRoot {
			return
		}
		if leftRoot < rightRoot {
			parent[rightRoot] = leftRoot
		} else {
			parent[leftRoot] = rightRoot
		}
	}

	serviceDomains := make(map[string][]string, len(serviceIDs))
	for _, serviceID := range serviceIDs {
		serviceDomains[serviceID] = routingDomains(serviceByID[serviceID].Domains)
	}
	for leftIndex, leftID := range serviceIDs {
		for _, rightID := range serviceIDs[leftIndex+1:] {
			if domainsOverlap(serviceDomains[leftID], serviceDomains[rightID]) {
				union(leftID, rightID)
			}
		}
	}

	components := make(map[string][]string)
	for _, serviceID := range serviceIDs {
		root := find(serviceID)
		components[root] = append(components[root], serviceID)
	}

	for _, members := range components {
		if len(members) < 2 {
			continue
		}
		sort.Strings(members)
		winner := ""
		for _, serviceID := range members {
			if normalized[serviceID] != "" && previous[serviceID] != normalized[serviceID] {
				winner = normalized[serviceID]
				break
			}
		}
		if winner == "" {
			winner = normalized[members[0]]
		}
		for _, serviceID := range members {
			if normalized[serviceID] != winner {
				normalized[serviceID] = winner
				changed = true
			}
		}
	}
	return normalized, changed
}

func canonicalizeServiceRoutes(routes map[string]string, services []Service) (map[string]string, bool) {
	normalized := make(map[string]string, len(routes))
	ranks := make(map[string]int, len(routes))
	changed := false
	routeIDs := make([]string, 0, len(routes))
	for serviceID := range routes {
		routeIDs = append(routeIDs, serviceID)
	}
	sort.Strings(routeIDs)
	for _, serviceID := range routeIDs {
		proxyID := strings.TrimSpace(routes[serviceID])
		if strings.TrimSpace(serviceID) == "" || proxyID == "" {
			changed = true
			continue
		}
		canonicalID, rank, ok := canonicalServiceRouteID(serviceID, services)
		if !ok {
			changed = true
			continue
		}
		if canonicalID != serviceID {
			changed = true
		}
		if existingRank, exists := ranks[canonicalID]; !exists || rank < existingRank {
			normalized[canonicalID] = proxyID
			ranks[canonicalID] = rank
		} else if normalized[canonicalID] != proxyID {
			changed = true
		}
	}
	if len(normalized) != len(routes) {
		changed = true
	}
	return normalized, changed
}

func canonicalServiceRouteID(serviceID string, services []Service) (string, int, bool) {
	serviceID = strings.TrimSpace(serviceID)
	if serviceID == "" {
		return "", 0, false
	}
	aliasToID := make(map[string]string, len(services)*2)
	exactIDs := make(map[string]struct{}, len(services))
	for _, service := range services {
		if service.ID == "" {
			continue
		}
		aliasToID[service.ID] = service.ID
		exactIDs[service.ID] = struct{}{}
		for _, alias := range service.Aliases {
			if alias = strings.TrimSpace(alias); alias != "" {
				aliasToID[alias] = service.ID
			}
		}
	}
	if canonicalID, ok := aliasToID[serviceID]; ok {
		if _, exact := exactIDs[serviceID]; exact {
			return canonicalID, 0, true
		}
		return canonicalID, 1, true
	}

	legacyBase := legacyServiceIDBase(serviceID)
	if legacyBase == "" {
		return "", 0, false
	}
	matches := make(map[string]struct{})
	for _, service := range services {
		if service.ID == "" {
			continue
		}
		if serviceSlugMatches(legacyBase, slug(service.Name)) || serviceSlugMatches(legacyBase, legacyServiceIDBase(service.ID)) {
			matches[service.ID] = struct{}{}
			continue
		}
		for _, alias := range service.Aliases {
			if serviceSlugMatches(legacyBase, legacyServiceIDBase(alias)) {
				matches[service.ID] = struct{}{}
				break
			}
		}
	}
	if len(matches) != 1 {
		return "", 0, false
	}
	for canonicalID := range matches {
		return canonicalID, 2, true
	}
	return "", 0, false
}

func stalePreviousRouteIDs(previous map[string]string, services []Service) []string {
	if len(previous) == 0 {
		return nil
	}
	result := make([]string, 0)
	for serviceID := range previous {
		canonicalID, _, ok := canonicalServiceRouteID(serviceID, services)
		if !ok || canonicalID != serviceID {
			result = append(result, serviceID)
		}
	}
	sort.Strings(result)
	return result
}

func legacyServiceIDBase(serviceID string) string {
	serviceID = strings.Trim(serviceID, "-")
	if serviceID == "" {
		return ""
	}
	parts := strings.Split(serviceID, "-")
	last := parts[len(parts)-1]
	if len(last) == 8 && isLowerHex(last) && len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], "-")
	}
	return serviceID
}

func serviceSlugMatches(left, right string) bool {
	left = strings.Trim(left, "-")
	right = strings.Trim(right, "-")
	if left == "" || right == "" {
		return false
	}
	return left == right || strings.HasPrefix(left, right+"-") || strings.HasPrefix(right, left+"-")
}

func isLowerHex(value string) bool {
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}
