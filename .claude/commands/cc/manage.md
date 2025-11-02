---
description: "Manage Claude Code configuration - edit commands, agents, and CLAUDE.md files"
---

You are tasked with helping the user manage their Claude Code configuration files with a focus on quality, consistency,
and adherence to best practices.
Use AskUserQuestion to determine what the user wants to do:

Ask the the user "What would you like to do with Claude Code configuration?"

Options:

1. Commands - Create/edit slash commands
2. Agents - Create/edit specialized agents
3. Skills - Create/edit Agent Skills for extending capabilities
4. Settings - Edit settings, permissions, hooks, MCP servers
5. CLAUDE.md - Edit CLAUDE.md or CLAUDE.local.md orchestration rules
6. Optimize - Audit and improve configuration for token efficiency

Based on the user's response:

- Commands → Read `.claude/guides/claude-code/commands.md` +
  WebFetch https://docs.claude.com/en/docs/claude-code/slash-commands
- Agents → Read `.claude/guides/claude-code/agents.md` +
  WebFetch https://docs.claude.com/en/docs/claude-code/sub-agents
- Skills → Read `.claude/guides/claude-code/skills.md` + WebFetch https://docs.claude.com/en/docs/claude-code/skills
- Settings → Read `.claude/guides/claude-code/settings.md` +
  WebFetch https://docs.claude.com/en/docs/claude-code/settings
- CLAUDE.md → Read `.claude/guides/claude-code/architecture.md`
- Optimize → Read `.claude/guides/claude-code/architecture.md`

After loading appropriate guide(s):

1. Follow the guide's principles and patterns
2. Apply Core Principles: Separation of Concerns, No Duplication, Token Efficiency, Clarity, Holistic Impact
3. Test all bash commands before adding to configuration
4. Validate changes against quality checklist

**STYLE:** No emojis. Clear, professional, technical language only.

## Additional Resources

- MCP: https://docs.claude.com/en/docs/claude-code/mcp
