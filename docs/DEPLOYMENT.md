# Deployment (VPS)

## You need one domain

A single domain serves everything. The port decides which program answers:

| Visitor does                | Port               | Answered by                                    |
| --------------------------- | ------------------ | ---------------------------------------------- |
| `ssh rivaldo.dev`           | 22                 | the portfolio TUI                              |
| opens `https://rivaldo.dev` | 443 (80 redirects) | Caddy → landing page explaining how to connect |
| `curl rivaldo.dev`          | 80/443             | Caddy → a terminal card with the command       |
| you, managing the server    | 2222               | the VPS's OpenSSH (`ssh -p 2222 root@<ip>`)    |

You do **not** need a second domain. `www.rivaldo.dev` is a free subdomain of the
same domain and redirects to the bare domain automatically.

```
                      rivaldo.dev  (A record → VPS IP)
                               │
          ┌────────────────────┼─────────────────────┐
          │ :22                │ :80 / :443          │ :2222
          ▼                    ▼                     ▼
   portfolio container    caddy container       host OpenSSH
   (SSH TUI)              (HTTPS, certificates)  (admin login)
                               │
                               └──▶ portfolio:8080 (landing page)
```

## 1. DNS

At your domain registrar (or Cloudflare), create:

| Type | Name  | Value      |
| ---- | ----- | ---------- |
| A    | `@`   | `<vps-ip>` |
| A    | `www` | `<vps-ip>` |

> ⚠️ **Cloudflare users:** set both records to **DNS only (grey cloud)**. Cloudflare's
> orange-cloud proxy only carries web traffic, so `ssh rivaldo.dev` would stop working.
> Caddy provides HTTPS itself, so you don't lose anything.

Check it: `dig +short rivaldo.dev` should print your VPS IP.

## 2. Move the admin SSH login to port 2222

The portfolio must own port 22, so the VPS's own login moves.

> ⚠️ Keep your current SSH session open until this step is verified. If something
> goes wrong, use it (or Hostinger's browser console) to undo the change.

```sh
sudo sed -i 's/^#\?Port .*/Port 2222/' /etc/ssh/sshd_config
sudo ufw allow 2222/tcp && sudo ufw allow 22/tcp && sudo ufw allow 80/tcp && sudo ufw allow 443   # if ufw is on
sudo systemctl daemon-reload
sudo systemctl restart ssh.socket 2>/dev/null || sudo systemctl restart ssh
```

If Hostinger's panel firewall is enabled (hPanel → VPS → Firewall), allow TCP
**22, 80, 443, 2222** and UDP **443** there too.

**Verify from your Mac, in a new terminal:** `ssh -p 2222 root@<vps-ip>`. Continue only
once this works, and update any CI or deploy scripts that SSH into the server.

## 3. Install Docker (skip if present)

```sh
curl -fsSL https://get.docker.com | sh
```

## 4. Deploy

From your Mac:

```sh
rsync -av --exclude .ssh --exclude bin -e "ssh -p 2222" \
  ~/projects/ssh-portfolio/ root@<vps-ip>:/opt/ssh-portfolio/
```

On the VPS:

```sh
cd /opt/ssh-portfolio
echo "DOMAIN=rivaldo.dev" > .env
docker compose up -d --build
```

Caddy gets an HTTPS certificate on its first request. That needs DNS (step 1) to
already point at the server and ports 80/443 to be open.

## 5. Verify

```sh
ssh rivaldo.dev                 # the portfolio
curl rivaldo.dev                # terminal card
open https://rivaldo.dev        # landing page with fingerprint
docker compose logs -f          # on the VPS
```

The landing page shows the server's **real host key fingerprint**. Check that it
matches what `ssh` prints on first connect.

## Updating

Edit `content/portfolio.yaml` or the code, run `make test`, then:

```sh
rsync -av --exclude .ssh --exclude bin -e "ssh -p 2222" ~/projects/ssh-portfolio/ root@<vps-ip>:/opt/ssh-portfolio/
ssh -p 2222 root@<vps-ip> 'cd /opt/ssh-portfolio && docker compose up -d --build'
```

The host key (`portfolio-data` volume) and certificates (`caddy-data`) survive
rebuilds. **Back up the host key** if you ever move servers, or every returning
visitor sees a "host identification has changed" warning:

```sh
docker compose exec portfolio cat /data/host_ed25519 > host_ed25519.backup
```

## Troubleshooting

| Symptom                                                        | Likely cause                                                                              |
| -------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| `ssh rivaldo.dev` asks for a password or reaches the VPS shell | OpenSSH is still on 22 (step 2), or the container isn't running                           |
| `ssh` hangs, but HTTPS works                                   | Cloudflare proxy is on (use grey cloud), or port 22 is closed in the firewall             |
| Browser shows a certificate error                              | DNS isn't pointing at the VPS yet, or 80/443 are blocked. See `docker compose logs caddy` |
| `curl` shows HTML                                              | `curl` was given `-H "Accept: text/html"`; plain `curl` gets the card                     |

## Alternative: Fly.io

`fly.toml` is included and covers the SSH side. For the web page on Fly, add an
`[http_service]` with `internal_port = 8080`, and Fly handles HTTPS instead of Caddy.
