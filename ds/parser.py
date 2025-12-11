import requests
from bs4 import BeautifulSoup
import pandas as pd
from transliterate import translit
from tqdm import tqdm
import re
"""
    df
        columns:
            Name - teachers name
            Url - teachers page url
            Table - кафдера
            NameAbbr
            EngName
            EngNameAbbr
            LastName
            EngLastName
"""

class Extractor:
    def __init__(self, df: pd.DataFrame):
        self.df = df
        self.urls = df['Url'].tolist()
    
    def extract_section(self, soup: BeautifulSoup, title):
        """
            Ищет заголовок h3 с текстом title и возвращает список текстов <li>.
            Поддерживает как старую структуру (h3 -> ol -> li), так и новую (h3 -> div.toggle_container -> div.block -> ol -> li).
        """
        header = soup.find("h3", class_="trigger", string=title)
        if not header:
            header = soup.find("h2", class_="trigger", string=title)
            if not header:
                header = soup.find("h4", class_="trigger", string=title)
                if not header:
                    return []

        # Старая структура: h3 -> ol -> li
        ol = header.find_next("ol")
        if ol:
            items = ol.find_all("li")
            return [item.get_text(" ", strip=True) for item in items]

        # Новая структура: h3 -> div.toggle_container -> div.block -> ol -> li
        toggle = header.find_next_sibling("div", class_="toggle_container")
        if toggle:
            block = toggle.find("div", class_="block")
            if block:
                # Вариант 1: ol внутри block
                ol = block.find("ol")
                if ol:
                    items = ol.find_all("li", recursive=False)
                    result = []
                    for item in items:
                        # Если внутри li есть ol (вложенный список), рекурсивно собрать все li из вложенного ol
                        nested_ol = item.find("ol")
                        if nested_ol:
                            nested_items = nested_ol.find_all("li")
                            result.extend([nested_item.get_text(" ", strip=True) for nested_item in nested_items])
                        else:
                            result.append(item.get_text(" ", strip=True))
                    return result
                # Вариант 2: ol после текста внутри block (например, 'Последние 20 публикаций')
                for child in block.children:
                    if getattr(child, 'name', None) == 'ol':
                        items = child.find_all("li", recursive=False)
                        result = []
                        for item in items:
                            nested_ol = item.find("ol")
                            if nested_ol:
                                nested_items = nested_ol.find_all("li")
                                result.extend([nested_item.get_text(" ", strip=True) for nested_item in nested_items])
                            else:
                                result.append(item.get_text(" ", strip=True))
                        return result

        return []

    def extract(self):
        """
        Собирает публикации по каждому преподавателю и возвращает список словарей,
        где ключ — фамилия преподавателя, значение — список публикаций.
        """
        data = {}
        for idx, url in enumerate(tqdm(self.urls)):
            last_name = self.df.iloc[idx]['LastName'] if 'LastName' in self.df.columns else self.df.iloc[idx]['Name'].split()[0]
            pubs = []
            if not isinstance(url, str) or not url:
                data[last_name] = []
                continue

            response = requests.get(url)
            response.encoding = "utf-8"
            soup = BeautifulSoup(response.text, "html.parser")
            
            pubs.extend(self.extract_section(soup, "Научные труды"))
            pubs.extend(self.extract_section(soup, "Некоторые научные публикации"))
            pubs.extend(self.extract_section(soup, "Публикации"))
            pubs.extend(self.extract_section(soup, "Научные монографии"))
            pubs.extend(self.extract_section(soup, "Учебно-методические материалы"))

            data[last_name] = pubs

        self.data = data
        return data
    
    def last_names(self):
        """
        Возвращает словарь, где ключ — фамилия преподавателя,
        значение — массив всех фамилий, которые встречаются в публикациях этого преподавателя.
        self.data - {teacher: [publications]}
        """
        if not hasattr(self, 'data'):
            raise ValueError("Сначала вызовите extract(), чтобы заполнить self.data")

        # Собираем список всех фамилий (русские и английские) из self.df
        if 'LastName' in self.df.columns:
            ru_last_names = set(self.df['LastName'].astype(str))
        else:
            ru_last_names = set(name.split()[0] for name in self.df['Name'].astype(str))

        en_last_names = set()
        if 'EngLastName' in self.df.columns:
            en_last_names = set(self.df['EngLastName'].astype(str))
        # Можно добавить и из EngName, если нужно

        all_last_names = ru_last_names | en_last_names

        result = {}
        for last_name, pubs in self.data.items():
            found = set()
            for pub in pubs:
                # Ищем все слова с заглавной буквы (русские и английские фамилии)
                ru_words = re.findall(r'\b[А-ЯЁ][а-яё]+\b', pub)
                en_words = re.findall(r'\b[A-Z][a-z]+\b', pub)
                found.update(w for w in ru_words if w in all_last_names)
                found.update(w for w in en_words if w in all_last_names)
            result[last_name] = list(found)
        return result
    
    def make_relations(self):
        """
        Возвращает словарь, где ключ — фамилия преподавателя,
        значение — массив фамилий, которые есть и в self.df, и встречаются в публикациях этого преподавателя.
        """
        pass