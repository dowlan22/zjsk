import os
from flask import Flask, request, jsonify
import socks
import socket

app = Flask(__name__)

# 读取环境变量设置端口
port = int(os.getenv("PORT", 8080))

# 用于存储用户名和密码的字典
users = {}

# 用于验证用户的装饰器
def requires_auth(f):
    def wrapper(*args, **kwargs):
        auth = request.authorization
        if not auth or not check_auth(auth.username, auth.password):
            return jsonify({"error": "Authentication failed"}), 401
        return f(*args, **kwargs)
    return wrapper

# 设置用户名和密码
@app.route('/users/<username>', methods=['PUT'])
@requires_auth
def add_user(username):
    users[username] = request.json['password']
    return jsonify({"message": "User added successfully"})

# 删除用户
@app.route('/users/<username>', methods=['DELETE'])
@requires_auth
def delete_user(username):
    users.pop(username, None)
    return jsonify({"message": "User deleted successfully"})

# 获取所有用户
@app.route('/users', methods=['GET'])
@requires_auth
def get_users():
    return jsonify(users)

# HTTP/HTTPS代理请求处理
@app.route('/', defaults={'path': ''})
@app.route('/<path:path>', methods=['GET', 'POST'])
@requires_auth
def proxy(path):
    url = request.url.replace(request.host_url, '', 1)
    if request.is_secure:
        return requests.request(
            method=request.method,
            url=url,
            headers=request.headers,
            data=request.get_data(),
            cookies=request.cookies,
            allow_redirects=False,
            verify=False
        )
    else:
        return requests.request(
            method=request.method,
            url=url,
            headers=request.headers,
            data=request.get_data(),
            cookies=request.cookies,
            allow_redirects=False
        )

# SOCKS5代理请求处理
@app.route('/socks', methods=['GET', 'POST'])
@requires_auth
def socks_proxy():
    socks.set_default_proxy(
        socks.SOCKS5,
        request.json['host'],
        request.json['port'],
        username=request.authorization.username,
        password=request.authorization.password
    )
    socket.socket = socks.socksocket
    return proxy()

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=port, debug=True)
