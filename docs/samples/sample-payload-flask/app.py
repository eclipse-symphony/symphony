from flask import Flask, render_template, request
import requests
import os

app = Flask(__name__, static_folder='/')

@app.route('/')
def index():
    env_vars = os.environ
    return render_template('index.html', env_vars=env_vars)

@app.route('/topology', methods=['POST'])
def topology():
    signal = request.form.get('signal')
    target = request.form.get('target')
    symphony_solution = os.environ.get('SYMPHONY_SOLUTION', 'default')
    symphony_component = os.environ.get('SYMPHONY_COMPONENT', 'component-a')
    data = {
        'from': symphony_component,
        'to': target,
        'solution': symphony_solution,
        'data': signal
    }
    url = 'http://localhost:8088/v1alpha2/vis-client'
    response = requests.post(url, json=data)
    return 'OK'

@app.route('/file/<path:file_path>')
def show_file(file_path):
    try:
        safe_file_path = escape(file_path)
        with open(safe_file_path, 'r') as f:
            file_contents = f.read()
        return render_template('file.html', file_path=safe_file_path, file_contents=file_contents)
    except FileNotFoundError:
        return f'File not found: {safe_file_path}', 404
    
@app.route('/env/<env_var>')
def show_env_var(env_var):
    try:
       return os.environ[env_var] + '\n'
    except KeyError:
        return f'Environment variable not found: {env_var}', 404
    

    
if __name__ == '__main__':
    app.run(host='0.0.0.0')