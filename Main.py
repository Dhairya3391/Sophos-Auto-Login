import requests
import time
import sys
import select
import argparse
import tty
import termios
import subprocess

# Configuration
base_url = 'http://10.255.1.1:8090/'
login_url = f'{base_url}/login.xml'
logout_url = f'{base_url}/logout.xml'
username = 'your_username'  # Replace with your actual username
password = 'your_password'  # Replace with your actual password

def login_to_portal():
    """Log in to the captive portal."""
    payload = {
        'mode': '191',
        'username': username,
        'password': password,
        'a': '1',  # This can be any value, often a timestamp is used
        'producttype': '0'
    }

    headers = {
        'Content-Type': 'application/x-www-form-urlencoded',
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
    }

    try:
        response = requests.post(login_url, data=payload, headers=headers)
        if response.status_code == 200:
            print('Login successful!')
            print('Response:', response.text)
        else:
            print(f'Login failed with status code {response.status_code}')
            print('Response:', response.text)

    except Exception as e:
        print(f'An error occurred during login: {e}')

def logout_from_portal():
    """Log out from the captive portal."""
    payload = {
        'mode': '193',  # Mode 193 is typically used for logging out
        'username': username,
        'a': '1',  # Can be any value, often a timestamp is used
        'producttype': '0'
    }

    headers = {
        'Content-Type': 'application/x-www-form-urlencoded',
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
    }

    try:
        response = requests.post(logout_url, data=payload, headers=headers)
        if response.status_code == 200:
            print('Logout successful!')
            print('Response:', response.text)
        else:
            print(f'Logout failed with status code {response.status_code}')
            print('Response:', response.text)

    except Exception as e:
        print(f'An error occurred during logout: {e}')

def check_ping():
    """Ping 1.1.1.1 to check for internet connectivity."""
    try:
        result = subprocess.run(['ping', '-c', '1', '-W', '2', '1.1.1.1'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        return result.returncode == 0
    except Exception:
        return False

def parse_args():
    parser = argparse.ArgumentParser(description='Captive Portal Auto Login Script')
    group = parser.add_mutually_exclusive_group()
    group.add_argument('-f', '--forever', action='store_true', help='Run forever until user quits')
    group.add_argument('-m', '--minutes', type=int, help='Run for specified minutes')
    group.add_argument('-H', '--hours', type=int, help='Run for specified hours')
    return parser.parse_args()

if __name__ == '__main__':
    args = parse_args()
    if args.forever:
        delay = None
        print('Running forever. Press q/Q or Ctrl+C to logout and stop...')
    elif args.minutes is not None:
        delay = args.minutes * 60
        print(f'Running for {args.minutes} minutes. Press q/Q or Ctrl+C to logout and stop...')
    elif args.hours is not None:
        delay = args.hours * 3600
        print(f'Running for {args.hours} hours. Press q/Q or Ctrl+C to logout and stop...')
    else:
        delay = 3600
        print('No duration flag given. Defaulting to 60 minutes. Press q/Q or Ctrl+C to logout and stop...')

    login_to_portal()
    start_time = time.time()
    last_status_check = time.time()
    try:
        fd = sys.stdin.fileno()
        old_settings = termios.tcgetattr(fd)
        tty.setcbreak(fd)
        while True:
            now = time.time()
            # Periodically check internet access every 30 seconds
            if now - last_status_check > 30:
                if not check_ping():
                    print('\nPing failed. Attempting to re-login...')
                    login_to_portal()
                last_status_check = now
            if delay is not None:
                elapsed = time.time() - start_time
                remaining = delay - elapsed
                if remaining <= 0:
                    print('\nAuto logout timer expired.')
                    break
                print(f'Press q or Q to logout and stop. Time left: {int(remaining)} seconds', end='\r', flush=True)
            else:
                print('Press q or Q to logout and stop.', end='\r', flush=True)
            i, o, e = select.select([sys.stdin], [], [], 1)
            if i:
                user_input = sys.stdin.read(1)
                if user_input.lower() == 'q':
                    print('\nDetected q/Q. Logging out and exiting...')
                    logout_from_portal()
                    sys.exit(0)
            if delay is not None and (time.time() - start_time) >= delay:
                break
    except KeyboardInterrupt:
        print('\nKeyboardInterrupt detected. Logging out and exiting...')
        logout_from_portal()
        sys.exit(0)
    finally:
        termios.tcsetattr(fd, termios.TCSADRAIN, old_settings)
    logout_from_portal()
