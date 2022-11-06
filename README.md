# GO in your Makefile

## Installation
Latest version\
You must have go 1.19 installed
```console
go install github.com/Hand-of-Doom/gomake@v1.0.0
```

## How to use
Create `Gomakefile` in /path/to/folder\
Then `cd /path/to/folder` and run `gomake`
```console
[foo@bar sample]$ pwd
/home/foo/sample
[foo@bar sample]$ ls
Gomakefile
[foo@bar sample]$ cat Gomakefile
sample:
    echo "Hello from bash!"
    out=go {
        fmt.Println("Hello from go!")
    }
    echo "$out"
[foo@bar sample]$ gomake sample
go: creating new go.mod: module gomakegen
go: to add module requirements and sums:
	go mod tidy
./sampletarget
Hello from bash!
Hello from go!
```

## Using golang libraries
```makefile
sample:
    go {
        import "github.com/google/uuid"
        fmt.Println(uuid.NewString())
    }
```
Then run `gomake sample`

## Full bash support
You can use all the features of bash since targets are bash scripts\
e.g. you can use multiline targets
```makefile
sample:
    echo "Hello, Gomake!"
    echo "Next line"
    echo "Next line"
```

## How it works?
It's just Makefile preprocessor\
Gomakefile builds into a Makefile so you keep using make

## Problems
### Is it slow? 
It's definitely slower than make cause gomake builds each go statement every time\
I can work on optimization but for now this is not my primary goal
### Is it Makefile compatible?
No. Gomakefile targets are bash scripts\
It's not compatible with Makefile\
You will need to rewrite your code
