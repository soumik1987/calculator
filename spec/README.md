# Spec

The spec is what we go by. If the code and the spec disagree, either the code is wrong or the spec needs updating first. Never the other way round, i.e. don't change the code and leave the spec behind.

requirements.md has what it should do, with the exact error messages and examples. design.md has how it's built, why, and the rules for writing code here. tasks.md has the build order and whatever's still open.

## Changing something

1. Edit requirements.md or design.md. Commit the spec change on its own so the diff is clean.
2. Run the change prompt below.
3. Go through the diff and the tests yourself before merging.

If you're not sure whether something belongs in requirements or design, ask whether a user or API client would notice it. If yes, requirements. If it's only about how it's built, design.

## Prompts

To build it from scratch:

Read the spec folder. Build it in the order in tasks.md, one step at a time, tests first, following the rules in design.md. Show me the tests passing after each step before moving on. If something in the spec is unclear or contradicts itself, ask me, don't guess.

To apply a change:

I changed the spec, look at the last commit on spec/. Work out what code and tests it affects. Update the tests first, then the code, then the READMEs. Don't touch anything the spec didn't ask for. If you think the spec is wrong tell me, don't work around it. Make sure make check and npm test pass.

To clear the open list:

Go through the Open section in spec/tasks.md one item at a time, tests first. Remove each line once it's done and everything passes.

To check for drift:

Go through requirements.md and check every rule against the code and tests. Tell me which rules have no test, and anything the code does that isn't in the spec. Don't fix anything yet.
