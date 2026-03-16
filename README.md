# my-skills

A command-line tool that scans a folder for skill definitions and generates a manifest file.

## Overview

Codex CLI with an ACP wrapper (e.g. [zed-industries/codex-acp](https://github.com/zed-industries/codex-acp), [cola-io/codex-acp](https://github.com/cola-io/codex-acp)) does not natively support custom skills. `my-skills` is a simple workaround to make custom skills available in an IDE via ACP and Codex — see [`AGENTS-example.md`](./AGENTS-example.md) for how to wire it up.

`my-skills` walks a configured folder, discovers `SKILL.md` files in its immediate subfolders, and writes a `skills.manifest` file summarizing each valid skill.

## Configuration

| Environment Variable | Required | Default | Description |
|---|---|---|---|
| `MY_SKILLS_PATH` | No | Current working directory | Path to the folder containing skill subfolders |

To persist it, add it to your shell config (e.g. `~/.bashrc`, `~/.zshrc`, `~/.profile`):

```sh
export MY_SKILLS_PATH=/path/to/skills
```

## Usage

```sh
# Using the default path (current directory)
my-skills

# Using a custom path
MY_SKILLS_PATH=/path/to/skills my-skills
```

No arguments needed.

## Output

The output file `skills.manifest` is written to `MY_SKILLS_PATH`. Each line is a compact JSON object representing one skill, with the following properties:

| Property | Description |
|---|---|
| `name` | The skill name from the `SKILL.md` frontmatter |
| `description` | The skill description from the `SKILL.md` frontmatter |
| `path` | The name of the subfolder containing the `SKILL.md` file |

## Validation

Frontmatter in each `SKILL.md` is strictly validated. Any file that fails validation is skipped and will not appear in the manifest. A warning is printed to stderr for each skipped file.

## AGENTS.md Integration

The repository includes an [`AGENTS-example.md`](./AGENTS-example.md) file showing how to use `my-skills` within an `AGENTS.md` workflow. It assumes the `my-skills` executable is on your `PATH` (which `go install` does by default). Append its contents to your project's `AGENTS.md` to get started.

## Requirements

- [Go](https://go.dev/dl/)

```sh
go install github.com/edmonl/my-skills@latest
```
