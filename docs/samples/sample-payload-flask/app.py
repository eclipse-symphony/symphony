from flask import Flask, render_template
import os

app = Flask(__name__, static_folder='/')

@app.route('/')
def index():
    env_vars = os.environ
    return render_template('index.html', env_vars=env_vars)

@app.route('/file/<path:file_path>')
def show_file(file_path):
    try:
        with open(file_path, 'r') as f:
            file_contents = f.read()
        return render_template('file.html', file_path=file_path, file_contents=file_contents)
    except FileNotFoundError:
        return f'File not found: {file_path}', 404
    
if __name__ == '__main__':
    app.run(host='0.0.0.0')