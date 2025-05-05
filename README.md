# sqlc-gen-better-python
A WASM plugin for SQLC allowing the generation of Python code.


> [!NOTE]  
> This is currently being worked on. It is far from being ready for any kind of release, let alone a stable one.  
> Please wait for the v1 release; before that, this plugin is likely to not work.

## Feature Support
Every [sqlc macro](https://docs.sqlc.dev/en/latest/reference/macros.html) is supported.
The supported [query commands](https://docs.sqlc.dev/en/latest/reference/query-annotations.html) depend on the SQL driver you are using, supported commands are listed below.
> Every `:batch*` command is not supported by this plugin and probably will never be.

> Prepared Queries are not planned for the near future, but will be implemented sooner or later

> [!NOTE]  
> Asyncpg only has very bad support until now. It doesn't support `:execresult`, `:execrows` and `:execlastid`

|           | `:exec` | `:execresult` | `:execrows` | `:execlastid` | `:many` | `:one` | `:copyfrom` |
|-----------|---------|---------------|-------------|---------------|---------|--------|-------------|
| aiosqlite | yes     | yes           | yes         | yes           | yes     | yes    | no          |
| sqlite3   | yes     | yes           | yes         | yes           | yes     | yes    | no          |
| asyncpg   | yes     | no            | no          | no            | yes     | yes    | no          |
| psycopg2  | no      | no            | no          | no            | no      | no     | no          |
| mysql     | no      | no            | no          | no            | no      | no     | no          |

## Development
A roadmap of what is planned & worked on can be found [here](https://github.com/users/rayakame/projects/1/)
### Changelog
Can be found [here](https://github.com/rayakame/sqlc-gen-better-python/blob/main/CHANGELOG.md)

## Credits
Because of missing documentation about creating these plugins, this work is heavily 
inspired by:
- [sqlc-gen-go](https://github.com/sqlc-dev/sqlc-gen-go)
- [sqlc-gen-java](https://github.com/tandemdude/sqlc-gen-java)

Special thanks to [tandemdude](https://github.com/tandemdude) for answering my questions on discord.