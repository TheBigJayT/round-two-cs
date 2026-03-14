#!/bin/bash

rm -f data/kills/*.pb
rm data/matches.jsonl
go build cmd/extract/main.go
./main -mode=extract -demo=$1
./main -mode=debug -demo=$1 > debug.txt
./main -list > read.txt
cat read.txt

