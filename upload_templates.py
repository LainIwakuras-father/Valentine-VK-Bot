import vk_api
import json
import os

# Настройки
GROUP_TOKEN = "vk1.a.q7B9lwEbJjBrPc2bwMr_GrbwLbxjoDjUBwDHiXPS4ToF6LGNOkD-H-1HHkdWaWVkojcp2fxdHk4N_aVE4MG6pkFDd0BT5TjrKUJo4HTfDAw_s9mkuQL0akgOSTeNTh5MhZ6qtPF0DbDBVmhJ9J9046VfcPiQaiD4t8Su1bnP8r37MDayh92JWGc3mo9WV3UptgfOrnjXSJumBxA1xeHaeQ"  # Вставьте ваш токен группы
GROUP_ID = -123456  # ID вашей группы с минусом (например: -123456)

vk_session = vk_api.VkApi(token=GROUP_TOKEN)
vk = vk_session.get_api()
upload = vk_api.VkUpload(vk_session)

def upload_template_photo(photo_path):
    """Загрузка фото для шаблона"""
    try:
        # Загружаем фото в альбом группы
        album_id = None  # Можно указать ID альбома или оставить None для загрузки в стену
        
        # Или загружаем в фото сообщений
        photo = upload.photo_messages(photo_path)[0]
        
        owner_id = photo['owner_id']
        photo_id = photo['id']
        
        return f"photo{owner_id}_{photo_id}"
    except Exception as e:
        print(f"Ошибка загрузки {photo_path}: {e}")
        return None

def add_templates_from_folder(folder_path):
    """Добавление шаблонов из папки"""
    templates = []
    
    # Список названий шаблонов
    template_names = [
        "Сердце с цветами",
        "Мишки с валентинкой",
        "Романтическая надпись",
        "Анимационная валентинка",
        "Милое сердечко"
    ]
    
    # Получаем список файлов в папке
    files = [f for f in os.listdir(folder_path) if f.lower().endswith(('.jpg', '.jpeg', '.png'))]
    files.sort()  # Сортируем по имени
    
    for i, filename in enumerate(files[:5]):  # Первые 5 файлов
        if i < len(template_names):
            photo_path = os.path.join(folder_path, filename)
            print(f"Загрузка: {filename} -> {template_names[i]}")
            
            attachment = upload_template_photo(photo_path)
            
            if attachment:
                templates.append({
                    "id": i + 1,
                    "name": template_names[i],
                    "attachment": attachment
                })
                print(f"✅ Успешно: {template_names[i]}")
            else:
                print(f"❌ Ошибка: {template_names[i]}")
    
    # Сохраняем в файл
    with open('templates.json', 'w', encoding='utf-8') as f:
        json.dump({"templates": templates}, f, ensure_ascii=False, indent=2)
    
    print(f"\n✅ Загружено {len(templates)} шаблонов")
    return templates

if __name__ == "__main__":
    # Путь к папке с шаблонами
    templates_folder = "шаблоны_валентинок"
    
    if os.path.exists(templates_folder):
        add_templates_from_folder(templates_folder)
    else:
        print(f"Создайте папку '{templates_folder}' и добавьте туда изображения шаблонов")
        os.makedirs(templates_folder, exist_ok=True)
        print(f"Папка создана. Добавьте туда изображения и запустите скрипт снова.")