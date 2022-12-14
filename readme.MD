# Usage

Check for status of platform with name `plat-stg-us-east-1-3`

```
$ CRDrill  --name plat-stg-us-east-1-1 --type platforms
```

```
$ CRDrill  --name company-a --type tenants
```


NOTE: THis sccript uses the current active context on kubeconfig file. By default it uses `~/.kube/config`, but you can specify an override value via env var:

```
$ export KUBECONFIG=~/.kube/<MY_CONFIG>
```

# How to build

## Makefile

You make use of the makefile that builds the tool automatically and places it on `/usr/local/bin`:

```
$ sudo make install
```

## Manually
1 - Clone the repo

```
$ git clone git@github.com:nm-morais/CRDrill.git
```

2 - Cd into it

```
$ cd CRDrill
```

3 - Run the command:

```
$ go build .
```

4 - This creates an executable with name `CRDrill`. Make executable with:

```
$ sudo chmod +x CRDrill
```

5 - Finally, to make it available in your path, copy the binary into one of your `bin` directories (we used `/usr/local/bin`) as an example:

```
$ sudo mv CRDrill /usr/local/bin
```

6 - Use the tool! `CRDrill  --name company-a --type tenants`


