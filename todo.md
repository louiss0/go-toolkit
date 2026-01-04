# Handoff

## Tests that are failing

None. `go test ./...` passes.

## What bugs are present

None known.

## What to do next

- Document the new init/config prompt flows and JSON summaries in `README.adoc`.
- If needed, add integration coverage for prompt-driven `init` in production mode via `-ldflags "-X .../mode.buildMode=production"`.
