# Deployment (VPS)

## One subdomain, SSH only

The website lives at `rvldodo.cloud` (hosted elsewhere). This project only serves
SSH, on a subdomain pointed at the VPS:

| Visitor does                      | Port | Answered by                                  |
| --------------------------------- | ---- | -------------------------------------------- |
| `ssh portfolio.rvldodo.cloud`     | 22   | the portfolio TUI                            |
| you, managing the server          | 2222 | the VPS's OpenSSH (`ssh -p 2222 root@<ip>`)  |

```
          portfolio.rvldodo.cloud  (A record → VPS IP)
                     │
          ┌──────────┴──────────┐
          │ :22                 │ :2222
          ▼                     ▼
   portfolio container     host OpenSSH
   (SSH TUI)               (admin login)
```

## 1. DNS

At your DNS provider (or Cloudflare), create:

| Type | Name        | Value      |
| ---- | ----------- | ---------- |
| A    | `portfolio` | `<vps-ip>` |

> ⚠️ **Cloudflare users:** set the record to **DNS only (grey cloud)**. Cloudflare's
> orange-cloud proxy only carries web traffic, so `ssh portfolio.rvldodo.cloud` would
> stop working.

Check it: `dig +short portfolio.rvldodo.cloud` should print your VPS IP.

## 2. Move the admin SSH login to port 2222

The portfolio must own port 22, so the VPS's own login moves.

> ⚠️ Keep your current SSH session open until this step is verified. If something
> goes wrong, use it (or Hostinger's browser console) to undo the change.

```sh
sudo sed -i 's/^#\?Port .*/Port 2222/' /etc/ssh/sshd_config
sudo ufw allow 2222/tcp && sudo ufw allow 22/tcp   # if ufw is on
sudo systemctl daemon-reload
sudo systemctl restart ssh.socket 2>/dev/null || sudo systemctl restart ssh
```

If Hostinger's panel firewall is enabled (hPanel → VPS → Firewall), allow TCP
**22** and **2222** there too.

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
docker compose up -d --build
```

## 5. Verify

```sh
ssh portfolio.rvldodo.cloud     # the portfolio
docker compose logs -f          # on the VPS
```

## Updating

Edit `content/portfolio.yaml` or the code, run `make test`, then:

```sh
rsync -av --exclude .ssh --exclude bin -e "ssh -p 2222" ~/projects/ssh-portfolio/ root@<vps-ip>:/opt/ssh-portfolio/
ssh -p 2222 root@<vps-ip> 'cd /opt/ssh-portfolio && docker compose up -d --build'
```

The host key (`portfolio-data` volume) survives rebuilds. **Back up the host key** if you ever move servers, or every returning
visitor sees a "host identification has changed" warning:

```sh
docker compose exec portfolio cat /data/host_ed25519 > host_ed25519.backup
```

## Troubleshooting

| Symptom                                                        | Likely cause                                                                              |
| -------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| `ssh portfolio.rvldodo.cloud` asks for a password or reaches the VPS shell | OpenSSH is still on 22 (step 2), or the container isn't running |
| `ssh` hangs                                                    | Cloudflare proxy is on (use grey cloud), or port 22 is closed in the firewall |

## Alternative: Fly.io

`fly.toml` is included: `fly launch --no-deploy`, create the `data` volume,
`fly deploy`, then `fly ips allocate-v4` and point the `portfolio` A record at it.
