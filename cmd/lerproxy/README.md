# lerproxy

Command lerproxy implements https reverse proxy with automatic LetsEncrypt and your own own TLS
certificates for multiple hostnames/backends including a static filesystem directory, nostr
DNS verification [NIP-05](https://github.com/nostr-protocol/nips/blob/master/05.md) hosting.

## Install

	go install lerproxy.mleku.dev@latest

## Run

```
Usage: lerproxy.mleku.dev [--listen LISTEN] [--map MAP] [--rewrites REWRITES] [--cachedir CACHEDIR] [--hsts] [--email EMAIL] [--http HTTP] [--rto RTO] [--wto WTO] [--idle IDLE] [--cert CERT]

Options:
  --listen LISTEN, -l LISTEN
                         address to listen at [default: :https]
  --map MAP, -m MAP      file with host/backend mapping [default: mapping.txt]
  --rewrites REWRITES, -r REWRITES [default: rewrites.txt]
  --cachedir CACHEDIR, -c CACHEDIR
                         path to directory to cache key and certificates [default: /var/cache/letsencrypt]
  --hsts, -h             add Strict-Transport-Security header
  --email EMAIL, -e EMAIL
                         contact email address presented to letsencrypt CA
  --http HTTP            optional address to serve http-to-https redirects and ACME http-01 challenge responses [default: :http]
  --rto RTO, -r RTO      maximum duration before timing out read of the request [default: 1m]
  --wto WTO, -w WTO      maximum duration before timing out write of the response [default: 5m]
  --idle IDLE, -i IDLE   how long idle connection is kept before closing (set rto, wto to 0 to use this)
  --cert CERT            certificates and the domain they match: eg: mleku.dev:/path/to/cert - this will indicate to load two, one with extension .key and one with .crt, each expected to be PEM encoded TLS private and public keys, respectively
  --help, -h             display this help and exit
```

`mapping.txt` contains host-to-backend mapping, where backend can be specified
as:

* http/https url for http(s) connections to backend *without* passing "Host"
  header from request;
* host:port for http over TCP connections to backend;
* absolute path for http over unix socket connections;
* @name for http over abstract unix socket connections (linux only);
* absolute path with a trailing slash to serve files from a given directory;
* path to a nostr.json file containing a
  [nip-05](https://github.com/nostr-protocol/nips/blob/master/05.md) and
  hosting it at `https://example.com/.well-known/nostr.json`
* using the prefix `git+` and a full web address path after it, generate html
  with the necessary meta tags that indicate to the `go` tool when fetching
  dependencies from the address found after the `+`.
* in the launch parameters for `lerproxy` you can now add any number of `--cert` parameters with
  the domain (including for wildcards), and the path to the `.crt`/`.key` files:

      lerproxy.mleku.dev --cert <domain>:/path/to/TLS_cert

  this will then, if found, load and parse the TLS certificate and secret key if the suffix of
  the domain matches. The certificate path is expanded to two files with the above filename
  extensions and become active in place of the LetsEncrypt certificates

  > Note that the match is greedy, so you can explicitly separately give a subdomain
  certificate and it will be selected even if there is a wildcard that also matches.

# IMPORTANT

With Comodo SSL (sectigo RSA) certificates you also need to append the intermediate certificate 
to the `.crt` file in order to get it to work properly with openssl library based tools like 
wget, curl and the go tool, which is quite important if you want to do subdomains on a wildcard
certificate.

Probably the same applies to some of the other certificate authorities. If you sometimes get 
issues with CLI tools refusing to accept these certificates on your web server or other, this 
may be the problem.

## example mapping.txt

    nostr.example.com: /path/to/nostr.json
	subdomain1.example.com: 127.0.0.1:8080
	subdomain2.example.com: /var/run/http.socket
	subdomain3.example.com: @abstractUnixSocket
	uploads.example.com: https://uploads-bucket.s3.amazonaws.com
	# this is a comment, it can only start on a new line
	static.example.com: /var/www/
    awesome-go-project.example.com: git+https://github.com/crappy-name/crappy-go-project-name

Note that when `@name` backend is specified, connection to abstract unix socket
is made in a manner compatible with some other implementations like uWSGI, that
calculate addrlen including trailing zero byte despite [documentation not
requiring that](http://man7.org/linux/man-pages/man7/unix.7.html). It won't
work with other implementations that calculate addrlen differently (i.e. by
taking into account only `strlen(addr)` like Go, or even `UNIX_PATH_MAX`).

## systemd service file

```
[Unit]
Description=lerproxy

[Service]
Type=simple
User=username
ExecStart=/usr/local/bin/lerproxy.mleku.dev -m /path/to/mapping.txt -l xxx.xxx.xxx.xxx:443 --http xxx.xxx.xxx.6:80 -m /path/to/mapping.txt -e email@example.com -c /path/to/letsencrypt/cache --cert example.com:/path/to/tls/certs
Restart=on-failure
Wants=network-online.target
After=network.target network-online.target wg-quick@wg0.service

[Install]
WantedBy=multi-user.target
```

If your VPS has wireguard running and you want to be able to host services from the other end of
a tunnel, such as your dev machine (something I do for nostr relay development) add the
`wg-quick@wg0` or whatever wg-quick configuration you are using to ensure when it boots,
`lerproxy` does not run until the tunnel is active.

## privileged port binding

The simplest way to allow `lerproxy` to bind to port 80 and 443 is as follows:

    setcap 'cap_net_bind_service=+ep' /path/to/lerproxy.mleku.dev

## todo

- add url rewriting such as flipping addresses such as a gitea instance
  `example.com/gituser/reponame` to `reponame.example.com` by funneling all
  `example.com/gituser` into be rewritten to be the only accessible user account on the gitea
  instance. or for other things like a dynamic subscription based hosting service subdomain
  instead of path