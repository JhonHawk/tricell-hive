# Naming language

Read before naming a new identifier, file, route, key, table or field, in code, a spec or a plan, and when reviewing the names a plan or a candidate introduces. The global guidance's «Language» holds the invariant; this leaf holds its detail. A name written in another language is implemented as written, and renaming it later costs a migration.

## What is written in English

- Identifiers (variables, functions, types, constants), file and directory names, code comments and docstrings, test names, descriptions and assertion messages, log messages and the errors a developer reads.
- Navigation routes and URL segments with their route and query parameters, route files and folders, and navigation ids used as a segment, those of public pages included: a URL path is an identifier, not text for the user.
- i18n keys, API names (paths, fields, parameters, error codes), database objects, environment variables, feature flags, event names and CSS classes.
- Each such name a spec or a plan introduces, even when its prose is in another language.

## What follows the product or session language

Interface text and the values of an i18n dictionary, the errors an end user reads, customer-facing documentation, and test fixtures and seed data. Commit messages and branch names follow the repository's convention, English by default.

## How to name

- An i18n key is semantic and hierarchical, such as `session.expiringWarning`; do not derive it from its text or from an id in another language, such as `module.empresas`.
- Translate a business term by what a practitioner of the domain calls it, not word for word: `accountsReceivable` for *cartera*, not `portfolio`.
- Use one language inside one identifier, and inside each enum or dictionary the work creates: `getByYear`, not `getByAno`. A new name is English also when it joins an enum, a dictionary or a table that holds names in another language; renaming those is a migration, as the global «Language» sets.
- Keep as written the acronym or proper name of a regulation or institution, such as `rfc`, `curp` or `cfdi`, each value the global «Language» keeps for compatibility, and each name a written rule of the project sets in another language. Translate every other term, enum values included.
- Write names in ASCII, without accents or `ñ`.

## Who acts

- **Every reader** lists each name the plan or the diff introduces in another language that this leaf does not keep.
- **A reviewer, a verifier or a build child** reports them with file and line and changes nothing outside its task; a build child also returns to its parent each business term it translated.
- **The main thread:**
  - corrects each one inside the work's scope, and reports it as a finding instead when the correction changes a contract, leaves that scope or changes a name the user's request or the ticket states;
  - records each business term it translates for the first time as `term → code name` under `## Glossary` in the project's `CODING_STANDARDS.md`, when that file exists or the plan's `Locations` name it, as part of the work item, and otherwise lists the pair in its report;
  - names in its report the project rule that kept a name the work item added in another language, and each enum, dictionary or table where its English name now sits beside names in another language, so the user can decide to migrate them;
  - proposes a spelling check on identifiers, such as cspell with an English dictionary and the kept terms as its word list, when the candidate carries at least one such name and the project runs no such check. It asks the user before adding it; in unattended work it adds the check only when the user authorized that kind of change beforehand, and reports it otherwise.
