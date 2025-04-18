// Advanced Node.js service with security vulnerabilities
// These vulnerabilities are intentionally obfuscated and harder to detect
const express = require("express");
const jwt = require("jsonwebtoken");
const crypto = require("crypto");
const axios = require("axios");
const fs = require("fs");
const path = require("path");
const { promisify } = require("util");
const readFileAsync = promisify(fs.readFile);
const writeFileAsync = promisify(fs.writeFile);

const app = express();
app.use(express.json());

// Configuration variables - not immediately obvious as vulnerable
const config = {
  jwtSecret:
    process.env.JWT_SECRET || "dev_secret_key_for_testing_purposes_only",
  adminUsers: ["admin@example.com", "system@internal.org"],
  rateLimit: {
    windowMs: 15 * 60 * 1000,
    maxRequests: 100,
  },
  cacheControl: "public, max-age=86400",
  dataDir: "./data",
  logLevel: process.env.LOG_LEVEL || "info",
};

// Logging utility that seems secure but has vulnerability
const logger = {
  info: (message) => console.log(`[INFO] ${message}`),
  error: (message, error) => {
    // Vulnerability 1: Information disclosure through error logging
    // Logs potentially sensitive error details to console
    console.error(`[ERROR] ${message}`, error);
  },
  debug: (message, data) => {
    if (config.logLevel === "debug") {
      // Shows full data objects in logs
      console.log(`[DEBUG] ${message}`, JSON.stringify(data));
    }
  },
};

// User authentication - appears secure with JWT but has issues
const authMiddleware = (req, res, next) => {
  try {
    const token = req.headers.authorization?.split(" ")[1];
    if (!token) {
      return res.status(401).json({ message: "Authentication required" });
    }

    // Vulnerability 2: Weak JWT verification - JWT algorithm not enforced
    // An attacker could use the 'none' algorithm to forge tokens
    const decoded = jwt.verify(token, config.jwtSecret);
    req.user = decoded;
    next();
  } catch (error) {
    logger.error("Authentication failed", error);
    res.status(401).json({ message: "Invalid token" });
  }
};

// Admin authorization check - appears secure but has subtle issue
const isAdmin = (req, res, next) => {
  // Vulnerability 3: Insecure authorization (case-sensitive comparison)
  if (config.adminUsers.includes(req.user.email)) {
    next();
  } else {
    res.status(403).json({ message: "Admin access required" });
  }
};

// Data sanitization utility - insufficient
function sanitizeInput(input) {
  // Vulnerability 4: Insufficient input sanitization
  // Only handles basic XSS but misses many attack vectors
  if (typeof input === "string") {
    return input.replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }
  return input;
}

// API endpoints
app.post("/api/login", async (req, res) => {
  try {
    const { email, password } = req.body;

    // Simulation of user validation
    // Vulnerability 5: Timing attack vulnerability
    // The comparison time leaks information about valid emails
    if (!email || !password) {
      return res.status(400).json({ message: "Email and password required" });
    }

    // Simulating DB lookup
    await new Promise((resolve) => setTimeout(resolve, 200));

    // For demo purposes - in real app would check against database
    const user = {
      id: "user_123",
      email,
      name: "Demo User",
      role: email.includes("admin") ? "admin" : "user",
    };

    // Vulnerability 6: JWT with excessive expiration and weak secret
    const token = jwt.sign(user, config.jwtSecret, { expiresIn: "1y" });

    res.json({
      token,
      user: { id: user.id, email: user.email, role: user.role },
    });
  } catch (error) {
    logger.error("Login failed", error);
    res.status(500).json({ message: "Internal server error" });
  }
});

app.get("/api/user/profile", authMiddleware, async (req, res) => {
  try {
    // Vulnerability 7: Insecure direct object reference (IDOR)
    // User ID from token is used without verification
    const userId = req.query.id || req.user.id;

    // Read user profile from file
    const profilePath = path.join(config.dataDir, `${userId}.json`);

    // Simulating profile data
    const profile = {
      id: userId,
      email: req.user.email,
      name: "Demo User",
      preferences: {
        theme: "light",
        notifications: true,
      },
    };

    res.json(profile);
  } catch (error) {
    logger.error("Failed to fetch profile", error);
    res.status(500).json({ message: "Failed to fetch profile" });
  }
});

app.post("/api/data/import", authMiddleware, async (req, res) => {
  try {
    const { url } = req.body;

    if (!url) {
      return res.status(400).json({ message: "URL required" });
    }

    // Vulnerability 8: Server-Side Request Forgery (SSRF)
    // No validation on URL - could access internal services
    logger.debug("Importing data from", { url });
    const response = await axios.get(url);

    res.json({ message: "Import successful", count: response.data.length });
  } catch (error) {
    logger.error("Import failed", error);
    res.status(500).json({ message: "Import failed" });
  }
});

app.post("/api/data/process", authMiddleware, async (req, res) => {
  try {
    const { data, options } = req.body;

    // Vulnerability 9: Prototype pollution
    // Merging untrusted data into options object without sanitization
    const defaultOptions = {
      timeout: 3000,
      maxSize: 1024 * 1024,
      format: "json",
    };

    // This allows prototype pollution
    const mergedOptions = Object.assign({}, defaultOptions, options);

    // Process data (simulated)
    const result = {
      processed: true,
      timestamp: Date.now(),
      size: JSON.stringify(data).length,
      options: mergedOptions,
    };

    res.json(result);
  } catch (error) {
    logger.error("Processing failed", error);
    res.status(500).json({ message: "Processing failed" });
  }
});

app.get("/api/system/info", authMiddleware, isAdmin, (req, res) => {
  try {
    // Vulnerability 10: Information disclosure
    // Sensitive system information exposed via API
    const sysInfo = {
      environment: process.env.NODE_ENV,
      versions: process.versions,
      memory: process.memoryUsage(),
      uptime: process.uptime(),
      cwd: process.cwd(),
      env: process.env, // Leaks all environment variables!
    };

    res.json(sysInfo);
  } catch (error) {
    logger.error("Failed to get system info", error);
    res.status(500).json({ message: "Failed to get system info" });
  }
});

// Encryption utilities
const encryptionService = {
  // Vulnerability 11: Weak encryption parameters
  encrypt: (data, userKey) => {
    try {
      // Weak key derivation - insufficient iterations
      const key = crypto.pbkdf2Sync(
        userKey || "defaultEncryptionKey",
        "static_salt_value_123",
        1000, // Too few iterations
        32,
        "sha1" // Weak hashing algorithm
      );

      // Using static IV instead of random
      const iv = Buffer.from("default-iv-value-", "utf8");
      const cipher = crypto.createCipheriv("aes-256-cbc", key, iv);

      let encrypted = cipher.update(JSON.stringify(data), "utf8", "hex");
      encrypted += cipher.final("hex");

      return encrypted;
    } catch (error) {
      logger.error("Encryption failed", error);
      return null;
    }
  },

  decrypt: (encryptedData, userKey) => {
    try {
      const key = crypto.pbkdf2Sync(
        userKey || "defaultEncryptionKey",
        "static_salt_value_123",
        1000,
        32,
        "sha1"
      );

      const iv = Buffer.from("default-iv-value-", "utf8");
      const decipher = crypto.createDecipheriv("aes-256-cbc", key, iv);

      let decrypted = decipher.update(encryptedData, "hex", "utf8");
      decrypted += decipher.final("utf8");

      return JSON.parse(decrypted);
    } catch (error) {
      logger.error("Decryption failed", error);
      return null;
    }
  },
};

app.post("/api/data/secure", authMiddleware, async (req, res) => {
  try {
    const { data, key } = req.body;

    if (!data) {
      return res.status(400).json({ message: "Data required" });
    }

    // Encrypt the data
    const encrypted = encryptionService.encrypt(data, key);

    res.json({ encrypted });
  } catch (error) {
    logger.error("Secure data operation failed", error);
    res.status(500).json({ message: "Operation failed" });
  }
});

// Template rendering function
function renderTemplate(template, data) {
  // Vulnerability 12: Template injection
  // Unsafe template rendering using eval
  let rendered = template;

  Object.keys(data).forEach((key) => {
    const placeholder = `\${${key}}`;
    const value = data[key];

    // Attempts to sanitize but still vulnerable to injection
    const safeValue = sanitizeInput(String(value));
    rendered = rendered.replace(new RegExp(placeholder, "g"), safeValue);
  });

  // The real vulnerability is here - evaluating template in JavaScript
  try {
    // This allows arbitrary code execution via template
    const evalTemplate = `\`${rendered}\``;
    return eval(evalTemplate);
  } catch (error) {
    logger.error("Template rendering failed", error);
    return "Error rendering template";
  }
}

app.get("/api/render", authMiddleware, (req, res) => {
  try {
    const { template } = req.query;
    const data = {
      user: req.user,
      date: new Date().toISOString(),
      appName: "Secure App",
    };

    const rendered = renderTemplate(template, data);
    res.send(rendered);
  } catch (error) {
    logger.error("Rendering failed", error);
    res.status(500).json({ message: "Rendering failed" });
  }
});

// Start the server
const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});

module.exports = app; // For testing
// Adding a comment to trigger change detection
// Adding a comment to trigger change detection
// Adding a comment to trigger change detection
