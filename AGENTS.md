# Maintenance

- Read the current README, version matrix, recent Git history and release changelogs before changing this project. Preserve the maintenance work recorded in earlier releases.
- The user requests subagents for this and future maintenance tasks. Delegate independent, bounded work with disjoint file ownership and review the results before integration.
- The primary control command is `dns`, implemented by `dns.sh`. Keep `prismctl.sh` as a compatibility entry. Public quick-start commands use Bash process substitution with the GitHub script URL.
- Preserve numeric menus, log options, saved credentials, transport mode, existing panel ports and upgrade rollback behavior.
- Do not modify or restart V2bX, XrayR, v2node, Docker, MTProxy or other business services when maintaining Prism.
- Test changed shell entry points and upgrade behavior before publishing. Keep generated installation commands and release script references consistent.
- Follow the user's current authorization for network connections and deployments. Never commit passwords, private keys, runtime databases, tokens or server-specific configuration.
