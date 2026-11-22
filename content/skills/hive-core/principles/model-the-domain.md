# Model the domain

Read when you write logic with state, or when a change would add one more branch to a growing chain of conditionals.

Put the domain rules in a data structure so the structure removes branches and invalid states:

- a state machine in place of several booleans or phase checks that must agree;
- a typed model in place of a long list of loose parameters;
- a table, registry or tagged union in place of the same branching repeated across files;
- a reducer in place of ad hoc mutations;
- an index, queue or normalized collection when the access pattern asks for one.

Signs you skipped it: a branch added to an if/else chain for the third time, or a second flag that must be kept in step with the first.

Do not force it. When the shape is clear, local and unlikely to grow, plain code is the better answer. An abstraction that adds a layer without removing a branch or an invalid state is not this principle.
