from datetime import datetime


def is_february_14():
    """Проверка, сегодня ли 14 февраля"""
    today = datetime.now()
    return today.month == 2 and today.day == 8
