Tool to improve db workflow.
	Currently generates helper structs for db column naming for defined structs.
	Tool uses source code ast tree parsing and generating source code accoding to defined template (text/Template).
```
To test 
go test -v

To run, there is example available in example dir
 ./dbhelper -tag db -structs User -path ./example/user.go
 ./dbhelper -tag json -path ./example/user.go -structs User,Person -suf dbcolum
 -path for source file
 -tag where column names defined
 -structs which structs to generate for sep by comma
 -suf suffix for generated file
```
