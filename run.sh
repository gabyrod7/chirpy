#!/bin/bash

go build -o out
timeout 30 "./out" &
