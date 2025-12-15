from abc import ABC, abstractmethod
from bs4 import BeautifulSoup
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from webdriver_manager.chrome import ChromeDriverManager
import time
import requests
from bs4 import BeautifulSoup
import pandas as pd
import re

class BaseParser(ABC):
    @abstractmethod
    def extract_publications(self, source):
        pass

class APMathParser(BaseParser):
    def extract_publications(self, soup: BeautifulSoup) -> list[str]:
        pubs = []

        for h in soup.find_all(["h2", "h3", "h4"], class_="trigger"):
            title = h.get_text(strip=True)

            if title not in {
                "Научные труды",
                "Некоторые научные публикации",
                "Публикации",
                "Научные монографии",
                "Учебно-методические материалы"
            }:
                continue

            ol = h.find_next("ol")
            if not ol:
                continue

            for li in ol.find_all("li"):
                pubs.append(li.get_text(" ", strip=True))

        return pubs


class SeleniumHTMLFetcher:
    _driver = None

    @classmethod
    def get_driver(cls):
        if cls._driver is None:
            options = Options()
            options.add_argument("--headless")
            options.add_argument("--no-sandbox")
            options.add_argument("--disable-gpu")
            options.add_argument("--window-size=1920,1080")

            cls._driver = webdriver.Chrome(
                service=Service(ChromeDriverManager().install()),
                options=options
            )
        return cls._driver

    @classmethod
    def get_html(cls, url: str) -> str:
        driver = cls.get_driver()
        driver.get(url)
        time.sleep(3)  # eLIBRARY грузит через JS
        return driver.page_source

class ELibraryParser(BaseParser):
    def extract_publications(self, url: str) -> list[str]:
        html = SeleniumHTMLFetcher.get_html(url)
        soup = BeautifulSoup(html, "html.parser")

        authors = []

        rows = soup.find_all("tr", id=lambda x: x and x.startswith("arw"))
        for row in rows:
            td = row.find("td", align="left")
            if not td:
                continue

            i_tag = td.find("i")
            if not i_tag:
                continue

            text = i_tag.get_text(strip=True)
            if text:
                authors.append(text)

        return authors

class Extractor:
    def __init__(self, df: pd.DataFrame):
        self.df = df
        self.parsers = {
            "Url": APMathParser(),   # обычные сайты
            "SI": ELibraryParser()   # Science Index
        }

    def extract(self):
        data = {}

        for _, row in self.df.iterrows():
            last_name = row["LastName"]
            pubs = []

            for column, parser in self.parsers.items():
                url = row.get(column)
                if not isinstance(url, str) or not url:
                    continue

                try:
                    if column == "SI":
                        pubs.extend(parser.extract_publications(url))
                    else:
                        r = requests.get(url, timeout=15)
                        r.encoding = "utf-8"
                        soup = BeautifulSoup(r.text, "html.parser")
                        pubs.extend(parser.extract_publications(soup))
                except Exception as e:
                    print(f"[WARN] {last_name} {column}: {e}")

            data[last_name] = list(dict.fromkeys(pubs))

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

