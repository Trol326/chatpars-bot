# Discord chat parser bot

Простенький бот для парсинга чатов в json для обучения бота [markov-discord](https://github.com/claabs/markov-discord/). 
Создаёт папку result, в которую складывает json с названием вида `имя-канала(id-канала).json`. [Пример формата данных](https://github.com/claabs/markov-discord/blob/master/img/example-training.json).

## Команды
Префикс - `pr!`
- `ping` - pong
- `pars #канал` - запустить парсинг канала, можно писать сколько угодно каналов. Если у бота нет доступа к каналу, он его проигнорирует. Пример использования: `pr!pars #канал1 #канал2`
- `status` - написать статус активных/завершенных парсингов(планировалось писать вместо ??? примерное общее количество, но сейчас bot api дискорда нет такой возможности). Данные хранятся в памяти программы, при перезагрузке теряются. Пример:

![image](https://github.com/user-attachments/assets/d7ada49e-7c3b-4c70-9266-133e2280e8df)
