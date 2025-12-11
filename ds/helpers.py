import pandas as pd
from transliterate import translit

def format_name(full_name):
    parts = full_name.split()
    if len(parts) == 4:
        last, first, middle, _ = parts
        return f"{last} {first[0]}.{middle[0]}."
    if len(parts) == 3:
        last, first, middle = parts
        return f"{last} {first[0]}.{middle[0]}."
    elif len(parts) == 2:
        last, first = parts
        return f"{last} {first[0]}."
    else:
        return full_name

def take_last_name(full_name):
    parts = full_name.split()
    return parts[0]

def df_maker() -> pd.DataFrame:
    teachers = pd.read_csv('amcp_teachers.csv')

    teachers['NameAbbr'] = teachers['Name'].apply(format_name)

    names = []
    for name in teachers['Name']:
        names.append(translit(name, language_code='ru', reversed=True))
    teachers['EngName'] = names

    teachers['EngNameAbbr'] = teachers['EngName'].apply(format_name)
    teachers['LastName'] = teachers['Name'].apply(take_last_name)
    teachers['EngLastName'] = teachers['EngName'].apply(take_last_name)

    return teachers