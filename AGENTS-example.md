# My Skills

Run `my-skills` when:
- The user requests a skill that is not available in the current session, and `my-skills` has not been run yet in this session.
- The user explicitly asks to load or scan their skills.

## Output

`my-skills` prints the path of a manifest file to stdout.
Each line of the manifest is a JSON object with the fields: `name`, `description`, and `path`, where `path` is the name of a subfolder in the same directory as the manifest file.
Each such subfolder is a skill folder containing a `SKILL.md`, which is the entry document for that skill.

If `my-skills` prints nothing to stdout, an error occurred and the manifest may not have been created or updated. Errors and warnings are printed to stderr.

The manifest reflects the state of the skill folders at the time of the run.
Always use the output of the most recent run; discard any previously read manifest.

## Apply a Skill

You do not have to re-run `my-skills` everytime you search for a skill. Use the manifest.

If the whole manifest has already been read after the last running of `my-skills`, use it directly without re-reading.

If you have not read the whole manifest yet:
- If you have relevant keywords, filter the manifest by skill name or description to find candidates. If filtering returns no results, read the full manifest.
- If you have no keywords, read the full manifest.

Once you identify the target skill, read its `SKILL.md` to learn how to apply it.
