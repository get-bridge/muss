# v0.8 - Unreleased

- Build statically linked executables.
- Rename "service" muss concept to "module" to improve clarity.
  Show deprecation warnings when old names are used.

# v0.7 - 2020-02-28

- Allow including files into service definition configs
  with an `include` value of `file: path`.
- Allow each item in `secret_commands` to have its own `passphrase`
  (defaulting to the global `secret_passphrase`).
- Run secret setup commands sequentially
  so that user input can be entered one at a time.
- Add "cache" parameter to `secret_commands` to allow disabling caching
  or setting an explicit expiration duration.

# v0.6 - 2020-02-04

- Fix load order to ensure MUSS_FILE and MUSS_USER_FILE are respected.

# v0.5 - 2020-02-04

- Allow "compose_file" to specify an alternate target file.
- Warn when COMPOSE_FILE is set but does not contain muss target.
- Set COMPOSE_PROJECT_NAME from optional "project_name" in muss.yaml.
- Provide friendlier error messages for registry authentication errors.
- Improve error output and exit codes.

# v0.4 - 2020-01-08

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
