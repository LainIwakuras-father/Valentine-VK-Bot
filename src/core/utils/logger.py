import logging
import sys
from logging.handlers import RotatingFileHandler
from pathlib import Path
from typing import Optional


# Цвета для консоли
class ColoredFormatter(logging.Formatter):
    COLORS = {
        "DEBUG": "\033[36m",  # Cyan
        "INFO": "\033[92m",  # Green
        "WARNING": "\033[93m",  # Yellow
        "ERROR": "\033[91m",  # Red
        "CRITICAL": "\033[95m",  # Magenta
    }
    RESET = "\033[0m"

    def format(self, record):
        log_color = self.COLORS.get(record.levelname, self.RESET)
        record.levelname = f"{log_color}{record.levelname}{self.RESET}"
        return super().format(record)


class LoggerConfig:
    CONSOLE_FORMAT = "[%(asctime)s] %(levelname)s [%(name)s]: %(message)s"
    FILE_FORMAT = (
        "[%(asctime)s] %(levelname)s [%(name)s:%(funcName)s:%(lineno)d]: %(message)s"
    )
    LOG_DIR = Path("logs")
    LOG_FILE = LOG_DIR / "peresil.log"


def setup_logging(log_level: str = "INFO", log_dir: Optional[Path] = None) -> None:
    if log_dir:
        LoggerConfig.LOG_DIR = log_dir
        LoggerConfig.LOG_FILE = log_dir / "peresil_bot.log"
    # Создать директорию если её нет
    LoggerConfig.LOG_DIR.mkdir(exist_ok=True)
    # Получить root logger
    root_logger = logging.getLogger()
    root_logger.setLevel(getattr(logging, log_level))
    # Удалить старые обработчики (если были)
    for handler in root_logger.handlers[:]:
        root_logger.removeHandler(handler)
    # ====== КОНСОЛЬ ======
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setLevel(getattr(logging, log_level))
    console_formatter = ColoredFormatter(LoggerConfig.CONSOLE_FORMAT)
    console_handler.setFormatter(console_formatter)
    root_logger.addHandler(console_handler)
    # ====== ФАЙЛ (с ротацией) ======
    file_handler = RotatingFileHandler(
        LoggerConfig.LOG_FILE,
        encoding="utf-8",
    )
    file_handler.setLevel(logging.DEBUG)  # Файл логирует ВСЕ
    file_formatter = logging.Formatter(LoggerConfig.FILE_FORMAT)
    file_handler.setFormatter(file_formatter)
    root_logger.addHandler(file_handler)


def get_logger(name: str) -> logging.Logger:
    return logging.getLogger(name)
