---
name: assess-pr
description: Assess the defect risk of a GitHub pull request with DevLens and get a merge verdict. Use when the user asks how risky a PR is, whether it is ready to merge, or gives a PR link or owner/repo#number to review.
---

Assess the pull request the user named: $ARGUMENTS

1. Normalize the reference to a repository (`owner/name`) and a number. Accept a full GitHub URL (`https://github.com/owner/name/pull/123`), `owner/name#123`, or `owner/name 123`. If nothing was given, ask for it.
2. Delegate the assessment to the `devlens:risk-assessor` subagent, passing the repository and number.
3. Return the subagent's answer to the user as it is. Do not recompute or soften the verdict.
