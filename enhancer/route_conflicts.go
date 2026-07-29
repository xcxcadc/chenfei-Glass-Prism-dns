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
	normalized := cloneStringMap(current)
	if len(normalized) < 2 {
		return normalized, false
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

	changed := false
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
