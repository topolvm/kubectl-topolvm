[![Main](https://github.com/topolvm/kubectl-topolvm/workflows/Main/badge.svg)](https://github.com/topolvm/kubectl-topolvm/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/topolvm/kubectl-topolvm)](https://goreportcard.com/report/github.com/topolvm/kubectl-topolvm)

# kubectl-topolvm

**🚧 Under development**

The utility command for TopoLVM.

Install
-------

```console
$ cd [kubectl-topolvm repository path]
$ go install .
```

Usage
-----

```console
$ kubectl topolvm node
NODE                    DEVICECLASS     USED            AVAILABLE       USE%
topolvm-example-worker  ssd             0               20396900352     0%
topolvm-example-worker  00default       1073741824      20396900352     5%
```
