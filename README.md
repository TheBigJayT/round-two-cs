# Round Two CS

## BUILDING

```bash
go build cmd/extract/main.go
```

## CLI

The program makes use of flags to control execution.

### EXTRACTING KILLS

To extract kills from a demo, use `-mode=extract` and `-demo=/path/to/demo`

```bash
./main -mode=extract -demo=testdemo.dem
```

This creates .pb files in `data/kills`. There is a .pb file created for each player on each side (T or CT). For a game with 10 players that would produce 20 .pb files (10 players times 2 sides).

### READING KILLS

Reading is the default behaviour of the program, as such `-mode=read` is not absolutely necessary.

To read every the total kills of every file in `data/kills`:

```bash
./main
```

or

```bash
./main -mode=read
```
