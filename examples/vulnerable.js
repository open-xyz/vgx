// Example vulnerable JavaScript file for VGX demo
const express = require("express");
const app = express();
const fs = require("fs");
const crypto = require("crypto");

// Set up body parsing middleware
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Hardcoded credentials - Security vulnerability: Sensitive data exposure
const DB_USER = "admin";
const DB_PASSWORD = "super_secret_password123";
const API_KEY = "sk_test_abcdefghijklmnopqrstuvwxyz123456";

// SQL Injection vulnerability
app.get("/users", (req, res) => {
  const userId = req.query.id;
  const query = `SELECT * FROM users WHERE id = ${userId}`; // Direct injection vulnerability

  // Simulating database query
  console.log(`Executing query: ${query}`);
  res.send(`Query executed: ${query}`);
});

// Path traversal vulnerability
app.get("/download", (req, res) => {
  const filename = req.query.file;

  // Vulnerable to path traversal
  const filePath = `/app/files/${filename}`;

  fs.readFile(filePath, (err, data) => {
    if (err) {
      return res.status(404).send("File not found");
    }
    res.send(data);
  });
});

// Cross-site scripting (XSS) vulnerability
app.get("/search", (req, res) => {
  const query = req.query.q;

  // Vulnerable to XSS
  res.send(`
    <html>
      <head><title>Search Results</title></head>
      <body>
        <h1>Search Results for: ${query}</h1>
        <div id="results">No results found</div>
      </body>
    </html>
  `);
});

// Weak encryption implementation
function encryptData(data) {
  // Using a static initialization vector - vulnerability
  const iv = Buffer.from("0000000000000000");
  const key = Buffer.from("supersecretkey123", "utf-8");

  const cipher = crypto.createCipheriv("aes-128-cbc", key, iv);
  let encrypted = cipher.update(data, "utf8", "hex");
  encrypted += cipher.final("hex");

  return encrypted;
}

// Command injection vulnerability
app.get("/ping", (req, res) => {
  const host = req.query.host;

  // Vulnerable to command injection
  const cmd = `ping -c 4 ${host}`;

  require("child_process").exec(cmd, (error, stdout, stderr) => {
    res.send(stdout);
  });
});

// Insecure cookie setting
app.get("/login", (req, res) => {
  // Set cookie without secure flag
  res.cookie("sessionId", "123456", {
    httpOnly: false,
    secure: false,
  });

  res.send("Logged in");
});

// Start the server
app.listen(3000, () => {
  console.log("Server running on port 3000");
});
// Adding a comment to trigger change detection
