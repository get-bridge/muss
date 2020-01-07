# v0.4 - Unreleased

- Add --index argument to "attach" command for use when a service is scaled
  (similar to the "exec" command)

# v0.3 - 2020-01-07

- Add "attach" command
- Add "version" command
- Ensure containers are stopped when "up" is interrupted
- Add --no-status option for "up"

# v0.2 - 2019-12-18

- Enable command to add fixed status line to bottom of "up" output.
- Add start, stop, restart, and rm commands.
- Use existing docker-compose.yml if there are no service defs.
- Improve signal handling.

# v0.1 - 2019-11-04

Initial release
