# sqlc-gen-better-python
A WASM plugin for SQLC allowing the generation of Python code.


> [!NOTE]  
> This is currently being worked on. It is far from being ready for any kind of release, let alone a stable one.  
> Please wait for the first GitHub release; before that, this plugin is likely to not work.

## Feature Support
> Every `:batch*` command is not supported by this plugin and probably will never be.

|           | `:exec` | `:execresult` | `:execrows` | `:execlastid` | `:many` | `:one` | `:copyfrom` |
| --------- | ------- | ------------- | ----------- | ------------- | ------- | ------ | ----------- |
| AioSqlite | yes     | yes           | yes         | yes           | yes     | yes    | no          |
|           |         |               |             |               |         |        |             |


## Development
A roadmap of what is planned & worked on can be found [here](https://github.com/users/rayakame/projects/1/)