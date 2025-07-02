<h1 align="center">Posto</h1>

<p align="center">
  <img src="https://github.com/user-attachments/assets/a8276a67-9638-4d6d-b5ba-240f75dd9503" alt="Posto Screenshot" />
</p>

<p align="center"><em>A privacy-first blogging platform with end-to-end encryption, built in Go.</em></p>

<p align="center">
  ✍️ Simple Publishing · 🔐 Zero-Knowledge Encryption · 🌐 Self-Hosted · 🧱 Go + MySQL + NGINX
</p>

---

**Posto** is a full-stack blogging service built with privacy, security, and self expression in mind.

Unlike most platforms, Posto requires no personal information from its users. Just a distinct username and password are needed, and all private content is encrypted using per-user keys that never leave memory. Whether you're journaling for yourself or sharing posts publicly, Posto puts you in full control of your data.

🌐 **Live Site**: [https://postoblog.duckdns.org](https://postoblog.duckdns.org)

---

## 🚀 Features

### 📝 Full Post Management
- Create, edit, and delete your blog posts through a clean, user-friendly interface.
- Only post owners see "Edit" and "Delete" buttons. Options are available on the profile page and individual blog post pages.

### 💖 Likes and Comments
- Posts can be **liked** by logged-in users.
- Visitors can leave **comments** on any public post (requires login).

### 👤 Profiles and Following
- View any user’s public profile and posts.
- **Follow** other users with a single click.
- Your **Feed** shows the latest posts from users you follow.

### 🔓 Public + Private Posts
- Mark posts as **public** or **private**.
- Private posts are fully encrypted and accessible only by the creator, using field-level encryption with per-user keys.

### 🔐 Zero-Knowledge Encryption
- Per-user keys are derived from passwords using Argon2.
- Keys live only in memory during a session and are never stored or shared.
- Even server admins cannot decrypt private content.

### 💬 Authenticated Interactions
- All social features (likes, comments, follows) require login.
- Cookie-based session management powered by Gorilla Sessions.

### 🧠 No Personal Info Required
- No email, no phone number, just a username and password.
- Anonymity and privacy are built into the platform.

---

## 🧱 Tech Stack

| Layer        | Tech                           |
|--------------|--------------------------------|
| **Backend**  | Go (Gin)                       |
| **Frontend** | HTML templates (SSR)           |
| **Database** | MySQL (AWS RDS), SQLite (dev)  |
| **Auth**     | Gorilla Sessions + bcrypt      |
| **Crypto**   | Argon2 key derivation          |
| **Hosting**  | AWS EC2 + NGINX + Certbot      |
| **Domain**   | Duck DNS                       |

---

## 📦 Local Development

You are free to clone this codebase and run Posto locally as a secure, standalone blogging service. This setup is ideal for private journaling or personal documentation.
Minor code adjustments may be needed in main.go for local-only mode.

- Social features (likes, comments, follows) will not function unless the app is accessed by multiple users over a network.
- You must configure your own database (MySQL or SQLite) and provide the required `.env` settings. TLS configuration is also required unless you revert to http.
- All private posts remain fully encrypted and accessible only to you.

> Local usage is completely self-contained, and no personal data ever leaves your machine.

---

### 🔧 Prerequisites
- Go 1.20+
- MySQL or SQLite
- Git

### 🛠 Quick Start

```bash
# Clone the repo
git clone https://github.com/your-username/posto.git
cd posto

# Create a .env file in the project root

# Run the server
go run main.go
```
---

## 📁 Configuration

### Example `.posto.env`

```env
MYSQL_USER=root
MYSQL_PASSWORD=your-password
MYSQL_HOST=localhost
MYSQL_DB=posto
COOKIE_STORE_KEY=your-very-secret-cookie-key
```

> These variables are required to run the app. Ensure your MySQL instance is accessible and the credentials are correct.

---

## 🔐 Security Notes

- Sessions are configured with:
  - `HttpOnly` cookies
  - `Secure` flag (HTTPS only)
  - `SameSite=Strict`
  - 7-day expiration
- Rate limiting and suspicious IP blocking middleware included
- Sessions use strong cookie-based encryption and are not persisted
- HTTPS enforced using NGINX and Certbot with automatic SSL renewal

---

## 🌐 Hosting Setup

- **Domain**: Free via [Duck DNS](https://www.duckdns.org/)
- **SSL**: HTTPS via [Certbot](https://certbot.eff.org/)
- **Database**: Hosted on AWS RDS (MySQL)
- **Server**: AWS EC2 (free tier eligible)
- **Reverse Proxy**: NGINX forwards requests from ports 80/443 → Go app (port 8080)
- **Systemd service**: Used to manage backend availability with fallback maintenance page

---

## 🔍 Learn More

- 📖 [Posto: A Technical Analysis](https://postoblog.duckdns.org/blogpost/1)  
   A deep dive into Posto’s architecture, encryption model, session handling, and hosting stack.

- 📝 [Welcome to Posto](https://postoblog.duckdns.org/blogpost/2)  
   Introductory post covering the motivation, goals, and initial design decisions behind Posto.

---

## 🪪 License

This project is licensed under the **MIT License**.  
See [https://mit-license.org/](https://mit-license.org/) for details.
