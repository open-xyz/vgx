#!/usr/bin/env python3
# Example vulnerable Python file for VGX demo

import os
import sqlite3
import pickle
import base64
import subprocess
from flask import Flask, request, render_template_string, redirect

app = Flask(__name__)

# Hardcoded credentials - Security vulnerability
DATABASE_USER = "admin"
DATABASE_PASSWORD = "db_password_123"
API_SECRET = "sk_live_5678abcdefghijklmnopqrstuvwxyz"

# SQL Injection vulnerability
@app.route('/users')
def get_user():
    user_id = request.args.get('id')
    
    # Vulnerable to SQL injection
    conn = sqlite3.connect('database.db')
    cursor = conn.cursor()
    query = f"SELECT * FROM users WHERE id = {user_id}"
    cursor.execute(query)
    
    return str(cursor.fetchall())

# Command injection vulnerability
@app.route('/system')
def execute_command():
    command = request.args.get('cmd')
    
    # Vulnerable to command injection
    result = os.system(command)
    
    return f"Command executed with status: {result}"

# Insecure deserialization
@app.route('/pickle')
def load_data():
    data = request.args.get('data')
    
    # Vulnerable to insecure deserialization
    decoded = base64.b64decode(data)
    obj = pickle.loads(decoded)
    
    return str(obj)

# Path traversal vulnerability
@app.route('/read')
def read_file():
    filename = request.args.get('file')
    
    # Vulnerable to path traversal
    with open(f"./files/{filename}", 'r') as f:
        content = f.read()
    
    return content

# Cross-site scripting (XSS) vulnerability
@app.route('/page')
def render_page():
    name = request.args.get('name', '')
    
    # Vulnerable to XSS
    template = f'''
    <!DOCTYPE html>
    <html>
    <head>
        <title>Welcome</title>
    </head>
    <body>
        <h1>Hello, {name}!</h1>
        <p>Welcome to our website.</p>
    </body>
    </html>
    '''
    
    return render_template_string(template)

# Insecure direct object reference
@app.route('/account/<account_id>')
def get_account(account_id):
    # No authorization check, vulnerable to IDOR
    # In a real app, this would fetch from a database
    accounts = {
        '1': {'name': 'Admin User', 'balance': '$10,000', 'ssn': '123-45-6789'},
        '2': {'name': 'Regular User', 'balance': '$100', 'ssn': '987-65-4321'}
    }
    
    if account_id in accounts:
        return str(accounts[account_id])
    else:
        return "Account not found"

# Information disclosure through error messages
@app.route('/divide')
def divide():
    try:
        a = int(request.args.get('a', '0'))
        b = int(request.args.get('b', '0'))
        result = a / b
        return f"Result: {result}"
    except Exception as e:
        # Vulnerable to information disclosure
        return f"An error occurred: {str(e)}"

if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0', port=5000) 