hedonic-games/
│
├── 🔵 GO КОД (Эдик)
│   ├── main.go                      # главная функция: загрузить графы → запустить алгоритм → сохранить результаты
│   ├── graph.go                     # Graph: узлы, рёбра, соседи
│   ├── hedonic.go                   # HedonicGame: утилита, best response, partition
│   ├── nash.go                      # NashStabilityFinder: алгоритм (лучший ответ динамика)
│   ├── metrics.go                   # модулярность, силуэт, число сообществ
│   ├── loader.go                    # загрузка Karate Club, Caveman из NetworkX или файлов
│   └── export.go                    # JSON (для Камиля), CSV результатов
│
preprocessing/
│
├── 🟡 PYTHON КОД (Камиль + Данил)
│   ├── data_loader.py               # загрузка готовых датасетов (Karate, Caveman, SBM)
│   ├── data_converter.py            # Graph ↔ JSON ↔ CSV ↔ edgelist
│   ├── data_processor.py            # обработка промежуточных результатов
│   ├── pipeline.py                  # интеграция: Go → Python → анализ
│   ├── test_data.py                 # юнит-тесты для Python модулей
│   └── analysis.ipynb               # Jupyter: графики, анализ, статистика
│
data/
│
├── 📊 ДАННЫЕ
│   ├── data/                        # исходные датасеты (edgelist, GML)
│   │   ├── karate.edgelist
│   │   └── caveman.edgelist
│   └── results/                     # результаты работы
│       ├── karate.json              # для визуализации
│       ├── karate.html              # интерактивная визуализация
│       └── results.csv              # таблица результатов
documentation/
│
├── 📚 ДОКУМЕНТАЦИЯ
│   ├── README.md                    # как запускать и что это такое
│   ├── CONTRIBUTING.md              # как развивать проект
│   └── faq.md                       # частые вопросы
│
├── 🚀 СКРИПТЫ
│   └── demo.sh                      # запустить всё по порядку
│
├── go.mod
├── go.sum
└── .gitignore
