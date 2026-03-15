# My Skills

When the user asks to load or scan their skills (e.g. "load my skills", "scan my skills"), run `my-skills`. It prints the path to a manifest file — read that file. Each line is a JSON object with the following fields: `name`, `description`, and `path`.

The user may ask for a specific skill, like "load my skills to do this and that". In this case, if you have keywords likely to appear in the skill name or description, use grep to narrow down relevant skills from the manifest rather than reading the whole file.

If you have no clue what to search for, or no ideal skill is found, read the full manifest.

If the user asks to reload or rescan (e.g. "reload my skills"), run `my-skills` again, read the new manifest, then replace any previously read skills with the new results.
