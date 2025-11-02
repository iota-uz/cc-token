---
description: "Optimize token usage in a Claude Code configuration file by making actual edits"
argument-hint: "<file-path>"
model: sonnet
disable-model-invocation: true
---

Optimize token usage in $1 by analyzing and implementing improvements.

## Analysis Phase

1. Run `cc-token count --analyze $1` to get baseline metrics
2. Read the file to understand current content and structure
3. Identify optimization opportunities:
    - Verbose prose → concise bullet points/tables
    - Repeated phrases/patterns → streamlined language
    - Inline content → extract to `.claude/guides/` with references
    - Duplicate content from CLAUDE.md → remove or reference
    - Excessive examples → consolidate to essential cases
    - Redundant formatting → simplified markdown
    - Long paragraphs → structured lists

## Implementation Phase

1. Apply optimizations iteratively with verification after EACH change:
    - Make ONE targeted edit to reduce verbosity
    - Run `cc-token count --analyze $1` to verify token reduction
    - If tokens increased or stayed same, revert change and try different approach
    - If tokens decreased, proceed to next optimization
    - Repeat for each optimization type:
        * Consolidate repetitive content
        * Simplify formatting and structure
        * Extract reusable content to guides when beneficial
        * Preserve meaning and clarity while reducing tokens

2. Re-run `cc-token count --analyze $1` for final measurement

## Output Format

Show concrete changes made:

```
TOKEN OPTIMIZATION RESULTS
File: [path]
Before: X tokens | After: Y tokens | Saved: Z tokens (W% reduction)

OPTIMIZATIONS APPLIED

1. [Optimization type]: [description]
   Lines affected: [line ranges]
   Token savings: ~X tokens

2. [Optimization type]: [description]
   Lines affected: [line ranges]
   Token savings: ~X tokens

[Additional optimizations...]

SUMMARY
- Total edits: N changes
- Token reduction: X → Y (-Z tokens, W%)
- Estimated cost savings: ~$X.XX per invocation
```

## Guidelines

- **CRITICAL**: Verify token count after EVERY edit - never skip this step
- Revert any change that increases token count or provides no benefit
- Make ONE edit at a time, verify, then proceed to next
- Preserve all critical information and instructions
- Maintain file purpose and functionality
- Keep examples essential for understanding
- Ensure edits don't break references or syntax
- Prioritize high-impact optimizations first
