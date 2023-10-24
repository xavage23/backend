def bold_print(*args, **kwargs):
    """Bold print"""
    print("\033[1m", end="")
    print(*args, **kwargs, end="")
    print("\033[0m")

def yellow_print(*args, **kwargs):
    """Yellow print"""
    print("\033[93m", end="")
    print(*args, **kwargs, end="")
    print("\033[0m")

def red_print(*args, **kwargs):
    """Red print"""
    print("\033[91m", end="")
    print(*args, **kwargs, end="")
    print("\033[0m")

def green_print(*args, **kwargs):
    """Green print"""
    print("\033[92m", end="")
    print(*args, **kwargs, end="")
    print("\033[0m")

def debug_print(*args, **kwargs):
    """Dull gray print"""
    print("\033[90m", end="")
    print(*args, **kwargs, end="")
    print("\033[0m")

def text_strip(text: str) -> str:
    """Given a line of text, strip it of whitespace and comments"""
    return text.strip().split("#")[0].strip()
