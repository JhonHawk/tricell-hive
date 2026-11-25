# Gate questions

Read before asking the delivery mode, the code review, the security review or the validation stop.

The delivery mode, the code review, the security review and the validation stop repeat in every work item, so they take these templates, with the header in the session language in place of the ID. The recommended option goes first, as `A` with the `(★)`, and the others follow in the order shown here, since the user reads the first option as the recommendation; each template shows one possible recommendation. [Delivery questions](../../git/delivery-ask.md) sets when the first three are asked, which options appear and which one is recommended. `direct-base` is not offered, since a repository's `## Hive` key sets it, so the Delivery card stays within four options; [flow-verify](../../../flow-verify/SKILL.md) step 2 sets when validation is asked. In a card each becomes one question as [questions](../questions.md#channel) sets, such as `question: <Stage> · <ticket> · ❓ Delivery - How is this change delivered? Situation: <…>`.

```
❓ **Delivery** - **How is this change delivered?**

Situation: <what the change touches and whether it changes a screen, interface text or the steps of a flow, or the project requires human validation>

A - (★) **interactive**: I stop before the push for you to validate work that changes a screen, interface text or a flow, or that the project requires you to validate; other work is not stopped. Then push, pull request and merge.

B - **automatic**: review, push, pull request and merge without stopping.

C - **ready**: review, push, a pull request with its CI and reviews answered, left open for you to merge.

D - **hold**: no commit and no publication; the verified working tree is the result.

➡️ <letter>: <the reason>
```

```
❓ **Review** - **Which code review runs on the candidate before your validation or the push?**

Situation: <what the change touches; a review the repository already runs on every pull request, by name>

A - (★) **Native: <host> <command>**: <host>'s own review of the diff.

B - **Agent: hive-review-code**: an independent child reviews the diff against the request and the criteria.

➡️ <letter>: <the reason, and whether the project's preferred one runs here>
```

```
❓ **Security review** - **Does hive-review-security also run on this candidate?**

Situation: <the security boundary the change touches>

A - (★) **Yes**: a child reviews that boundary before your validation or the push.

B - **No**: only the code review runs.

➡️ A: <what the boundary exposes>
```

```
❓ **Validation** - **Do you approve the result?**

Situation: <what to try, the reviews still running, and that approval is subject to their findings>

A - (★) **Approve**: <what the delivery does next>.

B - **Request corrections**: tell me what to change; I fix it and ask again.

➡️ A: <what the walk and the checks showed>
```
