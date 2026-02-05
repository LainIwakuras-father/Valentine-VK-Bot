import vk_api
from vk_api.keyboard import VkKeyboard, VkKeyboardColor
from vk_api.longpoll import VkLongPoll, VkEventType
from vk_api.upload import VkUpload
from vk_api.utils import get_random_id
from datetime import datetime
import json
import os
import re
import requests
from io import BytesIO
import time

# ========== НАСТРОЙКИ ==========
GROUP_TOKEN = "vk1.a.q7B9lwEbJjBrPc2bwMr_GrbwLbxjoDjUBwDHiXPS4ToF6LGNOkD-H-1HHkdWaWVkojcp2fxdHk4N_aVE4MG6pkFDd0BT5TjrKUJo4HTfDAw_s9mkuQL0akgOSTeNTh5MhZ6qtPF0DbDBVmhJ9J9046VfcPiQaiD4t8Su1bnP8r37MDayh92JWGc3mo9WV3UptgfOrnjXSJumBxA1xeHaeQ"  # Вставьте ваш токен группы

# ========== ИНИЦИАЛИЗАЦИЯ ==========
vk_session = vk_api.VkApi(token=GROUP_TOKEN)
vk = vk_session.get_api()
upload = VkUpload(vk_session)
longpoll = VkLongPoll(vk_session)

# ========== ФАЙЛЫ ДЛЯ ХРАНЕНИЯ ==========
SENT_FILE = "sent_valentines.json"
RECEIVED_FILE = "received_valentines.json"
TEMPLATES_FILE = "templates.json"

# Загрузка данных
def load_data(filename):
    if os.path.exists(filename):
        try:
            with open(filename, 'r', encoding='utf-8') as f:
                return json.load(f)
        except:
            return {}
    return {}

# Сохранение данных
def save_data(data, filename):
    with open(filename, 'w', encoding='utf-8') as f:
        json.dump(data, f, ensure_ascii=False, indent=2)

# Загружаем существующие данные
sent_valentines = load_data(SENT_FILE)
received_valentines = load_data(RECEIVED_FILE)

# Загрузка шаблонов валентинок
templates = load_data(TEMPLATES_FILE)
if not templates:
    templates = {
        "templates": [
            {"id": 1, "name": "Сердце с цветами", "attachment": ""},
            {"id": 2, "name": "Мишки с валентинкой", "attachment": ""},
            {"id": 3, "name": "Романтическая надпись", "attachment": ""},
            {"id": 4, "name": "Анимационная валентинка", "attachment": ""},
            {"id": 5, "name": "Милое сердечко", "attachment": ""}
        ]
    }
    save_data(templates, TEMPLATES_FILE)

# ========== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ==========

def create_keyboard(buttons, one_time=False, inline=False):
    """Создание клавиатуры из списка кнопок"""
    keyboard = VkKeyboard(one_time=one_time, inline=inline)
    
    for i, row in enumerate(buttons):
        if isinstance(row, list):
            # Если это строка с несколькими кнопками
            for j, btn in enumerate(row):
                if btn['color'] == 'positive':
                    keyboard.add_button(btn['label'], color=VkKeyboardColor.POSITIVE)
                elif btn['color'] == 'negative':
                    keyboard.add_button(btn['label'], color=VkKeyboardColor.NEGATIVE)
                elif btn['color'] == 'primary':
                    keyboard.add_button(btn['label'], color=VkKeyboardColor.PRIMARY)
                elif btn['color'] == 'secondary':
                    keyboard.add_button(btn['label'], color=VkKeyboardColor.SECONDARY)
                
                # Добавляем линию после кнопки, если это не последняя кнопка в строке
                if j < len(row) - 1:
                    keyboard.add_line()
        else:
            # Если это одиночная кнопка
            btn = row
            if btn['color'] == 'positive':
                keyboard.add_button(btn['label'], color=VkKeyboardColor.POSITIVE)
            elif btn['color'] == 'negative':
                keyboard.add_button(btn['label'], color=VkKeyboardColor.NEGATIVE)
            elif btn['color'] == 'primary':
                keyboard.add_button(btn['label'], color=VkKeyboardColor.PRIMARY)
            elif btn['color'] == 'secondary':
                keyboard.add_button(btn['label'], color=VkKeyboardColor.SECONDARY)
        
        # Добавляем новую строку после каждой строки кнопок, кроме последней
        if i < len(buttons) - 1:
            keyboard.add_line()
    
    return keyboard.get_keyboard()

def send_message(user_id, message, keyboard=None, attachment=None):
    """Отправка сообщения пользователю"""
    try:
        params = {
            'user_id': user_id,
            'message': message,
            'random_id': get_random_id(),
        }
        
        if keyboard:
            params['keyboard'] = keyboard
        if attachment:
            params['attachment'] = attachment
            
        vk.messages.send(**params)
        return True
    except Exception as e:
        print(f"Ошибка отправки: {e}")
        return False

def extract_user_id(input_text):
    """Улучшенная функция извлечения ID пользователя из разных форматов"""
    text = input_text.strip()
    
    # 1. Если это просто цифры - возвращаем как ID
    if text.isdigit():
        return int(text)
    
    # 2. Если это короткая ссылка (id123456789)
    if text.lower().startswith('id'):
        numbers = text[2:]  # Убираем 'id'
        if numbers.isdigit():
            return int(numbers)
    
    # 3. Если это ссылка ВКонтакте (поддерживаем .ru и .com)
    patterns = [
        # Форматы с .ru
        r'vk\.ru/id(\d+)',
        r'https?://vk\.ru/id(\d+)',
        r'https?://m\.vk\.ru/id(\d+)',
        r'vk\.ru/(\d+)',
        r'https?://vk\.ru/(\d+)',
        
        # Форматы с .com (для совместимости)
        r'vk\.com/id(\d+)',
        r'https?://vk\.com/id(\d+)',
        r'https?://m\.vk\.com/id(\d+)',
        r'vk\.com/(\d+)',
        r'https?://vk\.com/(\d+)',
        
        # С коротким именем (.ru)
        r'vk\.ru/([a-zA-Z0-9_.]+)',
        r'https?://vk\.ru/([a-zA-Z0-9_.]+)',
        r'https?://m\.vk\.ru/([a-zA-Z0-9_.]+)',
        
        # С коротким именем (.com)
        r'vk\.com/([a-zA-Z0-9_.]+)',
        r'https?://vk\.com/([a-zA-Z0-9_.]+)',
        r'https?://m\.vk\.com/([a-zA-Z0-9_.]+)',
        
        # Без указания домена (просто короткое имя)
        r'^https?://([a-zA-Z0-9_.]+)$',
    ]
    
    for pattern in patterns:
        match = re.search(pattern, text)
        if match:
            extracted = match.group(1)
            if extracted.isdigit():
                return int(extracted)
            else:
                # Это короткое имя, пробуем получить ID через API
                try:
                    result = vk.utils.resolveScreenName(screen_name=extracted)
                    if result and result.get('type') == 'user':
                        return result['object_id']
                except Exception as e:
                    print(f"Ошибка при получении ID по короткому имени {extracted}: {e}")
                    return None
    
    # 4. Если ввод выглядит как короткое имя без ссылки
    if re.match(r'^[a-zA-Z0-9_.]+$', text) and not text.isdigit():
        try:
            result = vk.utils.resolveScreenName(screen_name=text)
            if result and result.get('type') == 'user':
                return result['object_id']
        except:
            pass
    
    # 5. Пробуем извлечь ID из любой строки (последняя попытка)
    numbers = re.findall(r'\d{5,}', text)
    if numbers:
        return int(numbers[0])
    
    return None

def get_user_info(user_id):
    """Получение информации о пользователе"""
    try:
        users = vk.users.get(user_ids=user_id, fields='first_name,last_name,can_write_private_message')
        return users[0]
    except:
        return None

def is_february_14():
    """Проверка, сегодня ли 14 февраля"""
    today = datetime.now()
    return today.month == 2 and today.day == 5

def can_send_message_to_user(user_id):
    """Проверка, можно ли отправить сообщение пользователю"""
    try:
        user_info = get_user_info(user_id)
        if not user_info:
            return False
        
        can_write = user_info.get('can_write_private_message', True)
        return can_write
        
    except Exception as e:
        print(f"Ошибка при проверке возможности отправки: {e}")
        return True

# ========== ОСНОВНЫЕ ФУНКЦИИ БОТА ==========

def create_main_keyboard():
    """Главная клавиатура"""
    buttons = [
        {'label': '💌 Отправить валентинку', 'color': 'positive'},
        {'label': '📤 Мои отправленные', 'color': 'primary'},
        {'label': '📥 Мои полученные', 'color': 'primary'},
        {'label': 'ℹ️ Помощь', 'color': 'secondary'}
    ]
    return create_keyboard(buttons)

def show_welcome(user_id):
    """Приветственное сообщение"""
    welcome_msg = (
        "💘 Добро пожаловать в Бота Валентинок!\n\n"
        "Отправляйте анонимные или подписанные валентинки своим друзьям!\n"
        "Все валентинки будут доставлены 14 февраля 🎁\n\n"
        "Выберите действие:"
    )
    send_message(user_id, welcome_msg, create_main_keyboard())

def show_help(user_id):
    """Показать справку"""
    help_text = (
        "💘 БОТ ДЛЯ ВАЛЕНТИНОК\n\n"
        "📋 КАК ЭТО РАБОТАЕТ:\n"
        "1. Выбираете 'Отправить валентинку'\n"
        "2. Указываете, анонимная ли она\n"
        "3. Выбираете получателя по ссылке ВК\n"
        "4. Выбираете тип валентинки:\n"
        "   • Заготовленная - красивые шаблоны от дизайнеров\n"
        "   • Собственная - ваш текст + ваше фото\n"
        "5. Отправляете валентинку!\n\n"
        "📅 ВАЖНО:\n"
        "• Получатель увидит валентинку 14 февраля\n"
        "• Анонимные валентинки не показывают отправителя\n"
        "• Можно отправлять сколько угодно валентинок!\n\n"
        "🔗 КАК УКАЗАТЬ ПОЛУЧАТЕЛЯ (используйте .ru):\n"
        "1. Просто ID (цифры): 123456789\n"
        "2. Ссылка: vk.ru/id123456789\n"
        "3. Ссылка: https://vk.ru/id123456789\n"
        "4. Короткое имя (если есть): vk.ru/username\n"
        "5. Или просто username (если знаете)\n\n"
        "⚠️ Примечание: Также поддерживаются старые ссылки .com"
    )
    send_message(user_id, help_text, create_main_keyboard())

# ========== ПРОЦЕСС ОТПРАВКИ ВАЛЕНТИНКИ ==========

def start_valentine_creation(user_id):
    """Начало создания валентинки - выбор анонимности"""
    user_states[user_id] = {'step': 'anonymous_choice'}
    
    buttons = [
        [{'label': 'Да, анонимная 🎭', 'color': 'primary'}],
        [{'label': 'Нет, подписанная 📝', 'color': 'primary'}],
        [{'label': '❌ Отмена', 'color': 'negative'}]
    ]
    
    keyboard = create_keyboard(buttons, one_time=True)
    send_message(user_id, 
                 "💘 Выберите тип валентинки:\n\n"
                 "🎭 Анонимная - получатель не узнает, кто отправил\n"
                 "📝 Подписанная - укажется ваше имя",
                 keyboard)

def process_anonymous_choice(user_id, text):
    """Обработка выбора анонимности"""
    if text.lower() in ['да, анонимная 🎭', 'да', 'анонимная', '🎭']:
        anonymous = True
        anonymous_text = "анонимная"
    elif text.lower() in ['нет, подписанная 📝', 'нет', 'подписанная', '📝']:
        anonymous = False
        anonymous_text = "подписанная"
    else:
        send_message(user_id, "❌ Пожалуйста, выберите вариант из кнопок.")
        return False
    
    user_states[user_id] = {
        'step': 'recipient_input',
        'anonymous': anonymous,
        'valentine_data': {
            'anonymous': anonymous,
            'from_id': user_id,
            'date': datetime.now().strftime("%d.%m.%Y %H:%M")
        }
    }
    
    # Даем примеры форматов ввода
    examples = (
        "👤 Теперь укажите получателя любым способом:\n\n"
        "📝 ПРИМЕРЫ ФОРМАТОВ (используйте .ru):\n"
        "1. Просто ID: 123456789\n"
        "2. Со ссылкой: vk.ru/id123456789\n"
        "3. С https: https://vk.ru/id123456789\n"
        "4. Короткое имя: vk.ru/username\n"
        "5. Без ссылки: username\n\n"
        "🔍 Как найти ID пользователя?\n"
        "• Перейдите на его страницу\n"
        "• В адресной строке браузера будет:\n"
        "  - vk.ru/id123456789 (цифры - это ID)\n"
        "  - или vk.ru/username (короткое имя)\n\n"
        "❌ Для отмены напишите 'отмена'"
    )
    
    send_message(user_id, f"✅ Выбрана {anonymous_text} валентинка.\n\n{examples}")
    return True

def process_recipient_input(user_id, text):
    """Обработка ввода получателя"""
    if text.lower() == 'отмена':
        cancel_creation(user_id)
        return
    
    recipient_id = extract_user_id(text)
    
    print(f"DEBUG: Введенный текст: '{text}' -> Извлеченный ID: {recipient_id}")
    
    if recipient_id is None:
        # Даем более конкретные рекомендации
        error_msg = (
            "❌ Не удалось распознать получателя.\n\n"
            "📋 Возможные причины:\n"
            "• Вы ввели несуществующий ID или имя\n"
            "• Опечатка в ссылке или имени\n"
            "• У пользователя закрыт профиль\n\n"
            "🔧 Как исправить:\n"
            "1. Откройте страницу человека в ВК\n"
            "2. Скопируйте ссылку из адресной строки\n"
            "3. Вставьте её сюда (используйте .ru)\n\n"
            "Примеры правильного ввода (.ru):\n"
            "• 123456789\n"
            "• vk.ru/id123456789\n"
            "• https://vk.ru/id123456789\n"
            "• vk.ru/durov (если есть короткое имя)\n\n"
            "⚠️ Также работают старые ссылки .com"
        )
        send_message(user_id, error_msg)
        return
    
    # Проверка, не пытается ли отправить себе
    if recipient_id == user_id:
        send_message(user_id, "❌ Нельзя отправить валентинку самому себе!")
        cancel_creation(user_id)
        return
    
    # Получаем информацию о пользователе
    recipient_info = get_user_info(recipient_id)
    if not recipient_info:
        send_message(user_id, "❌ Пользователь не найден. Проверьте ID или ссылку.")
        return
    
    # Проверяем, можно ли писать пользователю
    if not can_send_message_to_user(recipient_id):
        send_message(user_id, 
                     "❌ У этого пользователя закрыты личные сообщения.\n"
                     "Выберите другого получателя.")
        cancel_creation(user_id)
        return
    
    recipient_name = f"{recipient_info['first_name']} {recipient_info['last_name']}"
    
    # Обновляем состояние
    user_states[user_id]['step'] = 'valentine_type'
    user_states[user_id]['valentine_data']['to_id'] = recipient_id
    user_states[user_id]['recipient_name'] = recipient_name
    
    # Клавиатура выбора типа валентинки
    buttons = [
        [{'label': '🎨 Заготовленная', 'color': 'primary'}],
        [{'label': '✏️ Собственная', 'color': 'primary'}],
        [{'label': '❌ Отмена', 'color': 'negative'}]
    ]
    keyboard = create_keyboard(buttons, one_time=True)
    
    send_message(user_id,
                 f"✅ Получатель найден: [id{recipient_id}|{recipient_name}]\n\n"
                 "🎨 Теперь выберите вид валентинки:\n\n"
                 "🎨 Заготовленная - красивые шаблоны от наших дизайнеров\n"
                 "✏️ Собственная - ваш текст + ваше фото (при желании)",
                 keyboard)

def process_valentine_type(user_id, text):
    """Обработка выбора типа валентинки"""
    if text.lower() in ['🎨 заготовленная', 'заготовленная', 'шаблон']:
        user_states[user_id]['step'] = 'template_choice'
        show_templates(user_id)
    elif text.lower() in ['✏️ собственная', 'собственная', 'своя']:
        user_states[user_id]['step'] = 'custom_text'
        send_message(user_id,
                     "✏️ Напишите текст вашей валентинки:\n\n"
                     "💌 Можно использовать эмодзи: ❤️💘🥰😘\n"
                     "📝 Максимум 1000 символов\n"
                     "🖼️ После текста можно прикрепить фото (необязательно)\n\n"
                     "❌ Для отмены напишите 'отмена'")
    else:
        send_message(user_id, "❌ Пожалуйста, выберите вариант из кнопок.")

def show_templates(user_id):
    """Показать список шаблонов"""
    buttons = []
    
    # Добавляем шаблоны по 3 в строку
    templates_to_show = templates['templates'][:6]  # Показываем максимум 6 шаблонов
    
    # Создаем строки по 2-3 кнопки в каждой
    for i in range(0, len(templates_to_show), 3):
        row = []
        for j in range(3):
            if i + j < len(templates_to_show):
                template = templates_to_show[i + j]
                row.append({'label': f"{template['id']}. {template['name']}", 'color': 'primary'})
        if row:
            buttons.append(row)
    
    # Добавляем кнопку отмены отдельной строкой
    buttons.append([{'label': '❌ Отмена', 'color': 'negative'}])
    
    keyboard = create_keyboard(buttons, one_time=True)
    
    message = "🎨 Выберите шаблон валентинки:\n\n"
    for template in templates['templates'][:6]:
        message += f"{template['id']}. {template['name']}\n"
    
    send_message(user_id, message, keyboard)

def process_template_choice(user_id, text):
    """Обработка выбора шаблона"""
    if text.lower() == 'отмена' or '❌' in text:
        cancel_creation(user_id)
        return
    
    try:
        template_num = int(text.split('.')[0])
    except:
        send_message(user_id, "❌ Пожалуйста, выберите шаблон из списка (цифру).")
        return
    
    selected_template = None
    for template in templates['templates']:
        if template['id'] == template_num:
            selected_template = template
            break
    
    if not selected_template:
        send_message(user_id, "❌ Шаблон не найден. Выберите из списка.")
        return
    
    user_states[user_id]['valentine_data']['template_id'] = selected_template['id']
    user_states[user_id]['valentine_data']['template_name'] = selected_template['name']
    user_states[user_id]['valentine_data']['attachment'] = selected_template.get('attachment', '')
    
    user_states[user_id]['step'] = 'template_text'
    send_message(user_id,
                 f"✅ Выбран шаблон: {selected_template['name']}\n\n"
                 "✏️ Теперь напишите текст для валентинки:\n"
                 "💌 Можно использовать эмодзи\n"
                 "📝 Максимум 500 символов\n\n"
                 "❌ Для отмены напишите 'отмена'")

def process_template_text(user_id, text):
    """Обработка текста для шаблона"""
    if text.lower() == 'отмена':
        cancel_creation(user_id)
        return
    
    if len(text) > 500:
        send_message(user_id, "❌ Текст слишком длинный. Максимум 500 символов.")
        return
    
    user_states[user_id]['valentine_data']['text'] = text
    confirm_valentine(user_id)

def process_custom_text(user_id, text):
    """Обработка текста для собственной валентинки"""
    if text.lower() == 'отмена':
        cancel_creation(user_id)
        return
    
    if len(text) > 1000:
        send_message(user_id, "❌ Текст слишком длинный. Максимум 1000 символов.")
        return
    
    user_states[user_id]['valentine_data']['text'] = text
    user_states[user_id]['step'] = 'custom_photo'
    
    buttons = [
        [{'label': '📷 Прикрепить фото', 'color': 'primary'}],
        [{'label': '➡️ Пропустить (без фото)', 'color': 'secondary'}],
        [{'label': '❌ Отмена', 'color': 'negative'}]
    ]
    keyboard = create_keyboard(buttons, one_time=True)
    
    send_message(user_id,
                 "✅ Текст сохранен!\n\n"
                 "📷 Хотите прикрепить фото к валентинке?\n"
                 "• Отправьте фото как вложение\n"
                 "• Или нажмите 'Пропустить'",
                 keyboard)

def handle_photo_attachment(user_id, event):
    """Обработка прикрепленного фото"""
    try:
        if hasattr(event, 'attachments') and event.attachments:
            for attachment in event.attachments:
                if attachment['type'] == 'photo':
                    photo = attachment['photo']
                    owner_id = photo['owner_id']
                    photo_id = photo['id']
                    access_key = photo.get('access_key', '')
                    
                    attachment_str = f"photo{owner_id}_{photo_id}"
                    if access_key:
                        attachment_str += f"_{access_key}"
                    
                    user_states[user_id]['valentine_data']['attachment'] = attachment_str
                    confirm_valentine(user_id)
                    return True
    except Exception as e:
        print(f"Ошибка обработки фото: {e}")
    
    return False

def process_custom_photo(user_id, event):
    """Обработка фото для собственной валентинки"""
    text = event.text.strip().lower() if hasattr(event, 'text') else ""
    
    if text == 'пропустить (без фото)' or text == 'пропустить':
        confirm_valentine(user_id)
        return
    
    if hasattr(event, 'attachments'):
        if handle_photo_attachment(user_id, event):
            return
    
    send_message(user_id, 
                 "❌ Фото не найдено в сообщении.\n"
                 "Отправьте фото как вложение или нажмите 'Пропустить'.")

def confirm_valentine(user_id):
    """Подтверждение отправки валентинки"""
    data = user_states[user_id]['valentine_data']
    recipient_name = user_states[user_id]['recipient_name']
    
    message = "✅ Валентинка готова к отправке!\n\n"
    message += f"👤 Получатель: [id{data['to_id']}|{recipient_name}]\n"
    message += f"🎭 Тип: {'Анонимная 🎭' if data['anonymous'] else 'Подписанная 📝'}\n"
    
    if 'template_name' in data:
        message += f"🎨 Шаблон: {data['template_name']}\n"
    else:
        message += "🎨 Тип: Собственная валентинка\n"
    
    if 'text' in data:
        preview = data['text'][:100] + "..." if len(data['text']) > 100 else data['text']
        message += f"💌 Текст: {preview}\n"
    
    if 'attachment' in data and data['attachment']:
        message += f"🖼️ С фото: Да\n"
    
    message += f"📅 Дата отправки: {data['date']}\n\n"
    message += f"📬 Получатель увидит валентинку 14 февраля!"
    
    buttons = [
        [{'label': '✅ Отправить', 'color': 'positive'}],
        [{'label': '❌ Отменить', 'color': 'negative'}]
    ]
    keyboard = create_keyboard(buttons, one_time=True)
    
    user_states[user_id]['step'] = 'confirmation'
    send_message(user_id, message, keyboard)

def send_valentine_final(user_id):
    """Финальная отправка валентинки"""
    data = user_states[user_id]['valentine_data']
    
    user_key = str(user_id)
    if user_key not in sent_valentines:
        sent_valentines[user_key] = []
    sent_valentines[user_key].append(data)
    save_data(sent_valentines, SENT_FILE)
    
    recipient_key = str(data['to_id'])
    if recipient_key not in received_valentines:
        received_valentines[recipient_key] = []
    received_valentines[recipient_key].append(data)
    save_data(received_valentines, RECEIVED_FILE)
    
    sender_message = "✅ Валентинка успешно отправлена!\n\n"
    sender_message += f"👤 Получатель: [id{data['to_id']}|{user_states[user_id]['recipient_name']}]\n"
    sender_message += f"📅 Дата доставки: 14 февраля\n\n"
    sender_message += "💝 Спасибо, что делитесь любовью!"
    
    send_message(user_id, sender_message, create_main_keyboard())
    
    try:
        if not data['anonymous']:
            sender_info = get_user_info(user_id)
            sender_name = f"[id{user_id}|{sender_info['first_name']} {sender_info['last_name']}]" if sender_info else "Кто-то"
            notification = f"💘 Вам отправили валентинку от {sender_name}!"
        else:
            notification = "💘 Вам отправили анонимную валентинку!"
        
        notification += "\n\n📅 Вы сможете прочитать её 14 февраля!"
        send_message(data['to_id'], notification)
    except Exception as e:
        print(f"Не удалось отправить уведомление получателю: {e}")
    
    if user_id in user_states:
        del user_states[user_id]

def cancel_creation(user_id):
    """Отмена создания валентинки"""
    send_message(user_id, "❌ Создание валентинки отменено.", create_main_keyboard())
    if user_id in user_states:
        del user_states[user_id]

def show_sent_valentines(user_id):
    """Показать отправленные валентинки"""
    user_key = str(user_id)
    user_sent = sent_valentines.get(user_key, [])
    
    if not user_sent:
        send_message(user_id, "📭 Вы еще не отправляли валентинок.")
        return
    
    message = "📤 Ваши отправленные валентинки:\n\n"
    for i, val in enumerate(user_sent, 1):
        recipient_info = get_user_info(val['to_id'])
        recipient_name = f"{recipient_info['first_name']} {recipient_info['last_name']}" if recipient_info else "Пользователь"
        
        message += f"{i}. Для: [id{val['to_id']}|{recipient_name}]\n"
        message += f"   📅 {val['date']}\n"
        message += f"   {'🎭 Анонимно' if val.get('anonymous', False) else '📝 Подписано'}\n"
        if val.get('text'):
            preview = val['text'][:50] + "..." if len(val['text']) > 50 else val['text']
            message += f"   💌 {preview}\n"
        message += "\n"
    
    send_message(user_id, message, create_main_keyboard())

def show_received_valentines(user_id):
    """Показать полученные валентинки"""
    if not is_february_14():
        today = datetime.now()
        if today.month == 2 and today.day < 14:
            days_left = 14 - today.day
            message = f"📅 Доступ к полученным валентинкам откроется через {days_left} дней!\n"
        elif today.month == 2 and today.day > 14:
            message = "📅 Полученные валентинки можно было посмотреть только 14 февраля!\n"
        else:
            message = "📅 Полученные валентинки можно будет посмотреть 14 февраля!\n"
        
        message += f"Сегодня: {today.strftime('%d.%m.%Y')}"
        send_message(user_id, message, create_main_keyboard())
        return
    
    user_key = str(user_id)
    user_received = received_valentines.get(user_key, [])
    
    if not user_received:
        send_message(user_id, "📭 У вас пока нет полученных валентинок.")
        return
    
    message = "📥 Ваши полученные валентинки:\n\n"
    for i, val in enumerate(user_received, 1):
        if val.get('anonymous', False):
            sender = "🎭 Аноним"
        else:
            sender_info = get_user_info(val['from_id'])
            sender = f"[id{val['from_id']}|{sender_info['first_name']} {sender_info['last_name']}]" if sender_info else "Пользователь"
        
        message += f"{i}. От: {sender}\n"
        message += f"   📅 {val['date']}\n"
        if val.get('text'):
            message += f"   💌 {val['text']}\n"
        message += "\n"
    
    send_message(user_id, message, create_main_keyboard())

# ========== ОСНОВНОЙ ЦИКЛ БОТА ==========

user_states = {}

print("=" * 50)
print("🤖 БОТ ДЛЯ ВАЛЕНТИНОК ЗАПУЩЕН")
print(f"📅 Дата: {datetime.now().strftime('%d.%m.%Y %H:%M:%S')}")
print("=" * 50)

for event in longpoll.listen():
    if event.type == VkEventType.MESSAGE_NEW and event.to_me:
        user_id = event.user_id
        text = event.text.strip() if hasattr(event, 'text') else ""
        
        print(f"[{datetime.now().strftime('%H:%M:%S')}] #{user_id}: {text[:50]}...")
        
        if text.lower() in ['начать', 'старт', 'start', 'привет', 'бот', '/start']:
            show_welcome(user_id)
            continue
        
        if user_id in user_states:
            state = user_states[user_id]
            
            if text.lower() == 'отмена' or text.lower() == '❌ отмена':
                cancel_creation(user_id)
                continue
            
            if state['step'] == 'anonymous_choice':
                process_anonymous_choice(user_id, text)
            
            elif state['step'] == 'recipient_input':
                process_recipient_input(user_id, text)
            
            elif state['step'] == 'valentine_type':
                process_valentine_type(user_id, text)
            
            elif state['step'] == 'template_choice':
                process_template_choice(user_id, text)
            
            elif state['step'] == 'template_text':
                process_template_text(user_id, text)
            
            elif state['step'] == 'custom_text':
                process_custom_text(user_id, text)
            
            elif state['step'] == 'custom_photo':
                process_custom_photo(user_id, event)
            
            elif state['step'] == 'confirmation':
                if text.lower() == '✅ отправить' or text.lower() == 'отправить':
                    send_valentine_final(user_id)
                elif text.lower() == '❌ отменить' or text.lower() == 'отменить':
                    cancel_creation(user_id)
                else:
                    send_message(user_id, "❌ Пожалуйста, используйте кнопки.")
            
            continue
        
        if text.lower() == '💌 отправить валентинку' or text.lower() == 'отправить валентинку':
            start_valentine_creation(user_id)
        
        elif text.lower() == '📤 мои отправленные' or text.lower() == 'мои отправленные':
            show_sent_valentines(user_id)
        
        elif text.lower() == '📥 мои полученные' or text.lower() == 'мои полученные':
            show_received_valentines(user_id)
        
        elif text.lower() == 'ℹ️ помощь' or text.lower() == 'помощь':
            show_help(user_id)
        
        else:
            send_message(user_id,
                         "🤔 Я не понял вашу команду.\n"
                         "Используйте кнопки ниже или напишите 'Помощь'",
                         create_main_keyboard())