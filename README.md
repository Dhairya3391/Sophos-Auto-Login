# Sophos Auto Login

This project provides a simple Python script to automate login and logout for a Sophos captive portal, commonly used in institutional or enterprise networks.

## Features

- Automatically logs in to the Sophos captive portal.
- Supports running for a specified duration (in minutes or hours) or indefinitely.
- Allows manual logout by pressing `q`/`Q` or using `Ctrl+C`.

## Requirements

- Python 3.x
- `requests` library (install with `pip install requests`)

## Compiling to a Standalone Binary (macOS, Optimized)

You can compile `Main.py` into a highly optimized standalone executable using Nuitka. This reduces memory and CPU usage compared to running with Python directly.

### Steps

1. **Install Nuitka and Required Tools**

   ```sh
   python3 -m pip install --upgrade pip nuitka
   ```

2. **Compile the Script**

   Navigate to your project directory:

   ```sh
   cd /Users/dhairya/Development/Sophos-Auto-Login
   ```

   Then run:

   ```sh
   nuitka Main.py \
     --standalone \
     --onefile \
     --enable-plugin=upx \
     --assume-yes-for-downloads \
     --follow-imports \
     --remove-output \
     --nofollow-import-to=tkinter,test \
     --noinclude-pytest-mode=nofollow \
     --no-pyi-file \
     --lto=yes \
     --clang \
     --static-libpython=yes \
     --low-memory \
     --force-stdout-spec=yes \
     --enable-console
   ```

3. **(Optional) Strip the Binary**

   To further reduce the binary size:

   ```sh
   strip auto_login
   ```

4. **Run the Optimized Binary**

   ```sh
   ./auto_login
   ```

---

## Configuration

Edit the `Main.py` file and set your actual username and password at the top:

```python
username = 'your_username'  # Replace with your actual username
password = 'your_password'  # Replace with your actual password
```

## Usage

Run the script from the command line:

```bash
python Main.py [options]
```

### Options

- `-f`, `--forever` : Run the script indefinitely until you quit.
- `-m MINUTES`, `--minutes MINUTES` : Run for the specified number of minutes.
- `-H HOURS`, `--hours HOURS` : Run for the specified number of hours.

If no option is provided, the script defaults to running for 60 minutes.

### Example

Run for 2 hours:

```bash
python Main.py --hours 2
```

Run indefinitely:

```bash
python Main.py --forever
```

## Stopping the Script

- Press `q` or `Q` and hit Enter to logout and exit.
- Or press `Ctrl+C` to logout and exit.

## Disclaimer

- This script is for educational purposes. Use it responsibly and only on networks where you have permission.
