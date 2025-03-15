#!/usr/bin/bash
oapi-codegen -package oapi -generate types swagger.yaml > types.go
oapi-codegen -package oapi -generate std-http swagger.yaml > server.go
oapi-codegen -package oapi -generate client swagger.yaml > client.go
