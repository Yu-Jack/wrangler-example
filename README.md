## Introduction

This is demo project to use [rancher/wrangler](https://github.com/rancher/wrangler) to write controllers. Comparing the client-go and controller-runtime, it provide more convinent interface to write the controllers. Besides, you don't need to generate code with CLI. Instead, it encapsulate into code piece that you could run with golang.

## Bootstrap

Run `make all`.


## Leader Election

Like distributed lock, usually we hope there is one controller to monitor the resource to avoid data race or race condition. So, we could make use of wrangler `leader.RunOrDie` which uses kubernetes lease to achieve the distributed lock.

More example [here](./pkg/example_group_a.go#L68)
