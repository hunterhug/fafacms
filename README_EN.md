# FaFa CMS: a distributed social content management system written in Golang

> 🌐 [简体中文](README.md) · English

[![GitHub forks](https://img.shields.io/github/forks/hunterhug/fafacms.svg?style=social&label=Forks)](https://github.com/hunterhug/fafacms/network)
[![GitHub stars](https://img.shields.io/github/stars/hunterhug/fafacms.svg?style=social&label=Stars)](https://github.com/hunterhug/fafacms/stargazers)
[![GitHub last commit](https://img.shields.io/github/last-commit/hunterhug/fafacms.svg)](https://github.com/hunterhug/fafacms)
[![Go Report Card](https://goreportcard.com/badge/github.com/hunterhug/fafacms)](https://goreportcard.com/report/github.com/hunterhug/fafacms)
[![GitHub issues](https://img.shields.io/github/issues/hunterhug/fafacms.svg)](https://github.com/hunterhug/fafacms/issues)

FaFa is a front-end/back-end decoupled content-community system (CMS) written in `Golang`, built around content interaction and connecting like-minded people. Ultimate goal: a usable content-community product. Companion front-end repo: [hunterhug/fafafront](https://github.com/hunterhug/fafafront) (Vue 3).

## Features

1. **User registration**: fill in QQ, Weibo, email, self-intro, avatar, etc., receive the registration email and click to activate. Inactive users see "inactive" and cannot use the platform. After activation they can log in and comment. There is no account deactivation. Blacklisted (banned) users cannot do anything. Publishing content and creating nodes requires an admin-granted VIP role. Summary: inactive / normal / VIP / admin — only VIP can create content, and admins can access privileged routes.
2. **Super-admin privilege control**: admins assign users to groups; each group owns a set of admin route resources (change other users' passwords, view all articles, view user info, ban users, etc.). Without access to these special routes, a user only operates on their own resources; otherwise the corresponding group permission is required. Invisible to normal users.
3. **User profile operations**: log in/out anytime, supplement profile info, change password, reset password via email. Nickname can be changed at most twice per month and must be globally unique. Two-factor authentication (2FA, standard TOTP) binding & login, with an "Account Security" section in the personal center (view / set / disable 2FA); admins can help reset it.
4. **Content editing**: VIP users can create content nodes (at most two levels) with drag sorting. Under a node you can create articles, hide them, pin them, password-protect them, etc. Articles have a special publish mechanism with history versions: content is first saved to a pre-publish field, and only refreshes into the formal field when published; edits can be saved into history, each publish is recorded into publish history, and you can restore from any historical version. Articles support drag sorting and two-step deletion: moving to the recycle bin, then either restoring or permanently deleting (permanent delete is still a soft delete status=4 — gone from the user's world but still visible to admins, no physical delete, and it does not cascade-delete comments / likes / notifications).
5. **Homepage reading & comments**: all users can browse and comment on others' articles; owners can toggle comments on/off. Comments are QQ-Music-style; owners can delete their own comments. Users can like/unlike articles or comments (recorded to prevent duplicate likes), and report articles or comments. The server supports auto-ban and a report threshold (auto-ban when reports exceed the threshold). Comments have anti-spam captcha (triggered on high frequency). Admins have a report list to review reports and ban/unban.
6. **File storage**: avatars, node backgrounds, article backgrounds and other internal images must be uploaded via the upload API and stored in the database; unsafe external image links are forbidden. Files are stored locally or in cloud object storage, with list / classify / tag APIs.
7. **Server configuration**: registration can be disabled. Admin-privileged users can blacklist / unblacklist users, activate users, create users, ban content, and grant VIP.
8. **In-site messages**: comment liked, content liked, content commented, comment commented; system notifications for content banned, comment banned, and admin broadcast (fanned out to all users). Notification senders / articles / comments are clickable; deleted articles/comments show placeholders "Article deleted" / "Comment deleted".
9. **Follow users**: users can follow each other; when someone publishes content, their followers get an in-site notification.
10. **Private messages**: user-to-user chat. (addon)
11. **User & content keyword search**. (addon)
12. **Security**: request signing (HMAC, enforced by default); anti-brute-force + captcha for login/register/forget-password/change-password; anti-crawler rate limiting + captcha unblock; bcrypt password storage; two-factor auth (2FA); comment anti-spam captcha; security headers (CORS / XSS, etc.). Security state is stored in Redis for multi-replica deployment. Email verification codes are 6-digit; activation/reset codes are invalidated after 5 wrong attempts, with remaining attempts shown.
15. **Encryption at rest**: sensitive database fields (user profile, private messages, post bodies, comments, report reasons) and locally stored uploads are written encrypted, so a leaked dump, backup or stolen disk does not expose content. Keys use envelope encryption: a root key (KEK) from the environment or generated by the deploy script derives a per-user data key (DEK) and a per-file key (FEK); encryption binds table, column and owner, so moved ciphertext fails to decrypt. Fields that need equality lookups also store a blind index (an HMAC digest), which means **search is exact-match only, not fuzzy**. When the root key is absent the server falls back to a built-in default and warns in the startup log, so local development needs no configuration.
13. **Site config & friend links**: admins can customize the **site title, site subtitle, community intro and footer intro**; friend links support CRUD / sort / hide / open-in-new-window; the public API returns visible friend links.
14. **SEO URLs & node hiding**: a node page is `/<user>/node/<nodeSeo>` and an article page is `/<user>/node/<nodeSeo>/<articleSeo>`. Article slugs are unique inside their node while node slugs are unique per account, both generated randomly by default (still editable, with uniqueness checks); moving an article into a node that already uses its slug appends a suffix. A node can be hidden (from the author's node tree or the admin console): articles in it and in its child nodes leave the public lists and are visible only to their author, while direct links keep working.
15. **Mobile**: the frontend is fully mobile-adapted (responsive layout, mobile bottom navigation, fullscreen chat / editor, etc.).

## Product showcase

| Home feed | Article | User profile |
| --- | --- | --- |
| ![Home feed](./docs/screenshots/home.png) | ![Article](./docs/screenshots/content.png) | ![User profile](./docs/screenshots/profile.png) |

| Explore | Fans | Following |
| --- | --- | --- |
| ![Explore](./docs/screenshots/explore.png) | ![Fans](./docs/screenshots/fans.png) | ![Following](./docs/screenshots/follows.png) |

| Login | Personal center | Admin console |
| --- | --- | --- |
| ![Login](./docs/screenshots/login.png) | ![Personal center](./docs/screenshots/personal.png) | ![Admin console](./docs/screenshots/admin.png) |

| Notifications | Private chat |
| --- | --- |
| ![Notifications](./docs/screenshots/notifications.png) | ![Private chat](./docs/screenshots/chat.png) |

## Tech stack

| Layer | Tech |
| --- | --- |
| Backend | Go 1.23 · Gin · XORM · MySQL · Redis · Aliyun OSS · gomail (SMTP) |
| Frontend | Vue 3 · Vite · Element Plus (10 UI languages) |
| Deploy | Docker / docker compose one-click (nginx static hosting + reverse proxy), multi-replica capable |

## Directory layout

```
fafacms/
├── main.go              # entry (config / table init / routing)
├── core/
│   ├── config/          # global config & version
│   ├── controllers/     # business handlers (user/content/comment/message/file/report …)
│   ├── model/           # data models (xorm)
│   ├── router/          # route table (public + /v1 authed routes)
│   ├── server/          # gin setup & middlewares (sign/ratelimit/security headers)
│   ├── session/         # Redis session
│   ├── util/            # utilities (oss/mail/rdb/captcha/totp …)
│   └── flog/            # logging
├── docs/                # API docs (docsify) & screenshots
├── install/             # one-click deploy (deploy.sh + docker-compose)
├── VERSION_LOG.md       # release history
└── TODO_LIST.md         # development plan
```

## Quick start (Docker one-click)

Have a Unix-like machine with `Docker` and `docker compose` installed, then:

```bash
git clone https://github.com/hunterhug/fafacms
cd fafacms/install
./deploy.sh            # up/ps/logs/down/rebuild
```

The script brings up `mysql`, `redis`, `phpMyAdmin`, the backend API and the API docs, then prints the access URLs (including the LAN IP):

| Service | URL |
| --- | --- |
| Backend API | http://127.0.0.1:8080 |
| API docs | http://127.0.0.1:9889 |
| phpMyAdmin | http://127.0.0.1:8000 |

Data persistence lives outside the repo: macOS `~/.fafacms`, Linux `/opt/fafacms`, overridable with the `FAFACMS_DATA_DIR` env var; see `install/config.yaml`.

> Demo account: admin `admin/admin` (auto-created on first start — change it soon). Example passwords are for local demo only.

For frontend local development and integration, see the [hunterhug/fafafront](https://github.com/hunterhug/fafafront) README.

## Docs navigation

- [Release history](VERSION_LOG.md): version change records.
- [Development plan](TODO_LIST.md): feature plan and progress (done / planned).
- [API list](docs/http/api.md): API documentation index.
- [Deployment guide](docs/deploy.md): end-to-end deployment (server → backend → frontend → HTTPS).
- [Backend deployment](install/README.md): backend one-click deployment and configuration.

# License

Licensed under the [GNU Affero General Public License v3.0](https://www.gnu.org/licenses/agpl-3.0.html) (AGPL-3.0).

You are free to use, modify and distribute this project, including for commercial purposes. However, if you offer it to users over a network, AGPL-3.0 requires you to provide those users with the corresponding complete source code. See [LICENSE](LICENSE) for the full terms.

## Acknowledgments

- [Acknowledgments](ACKNOWLEDGMENTS.md): thanks to every developer who contributed.
