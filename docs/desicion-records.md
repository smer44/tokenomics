# DR001: Агентов-заказичики VS централизация

## Контекст

Если каждый агент будет принимать частное решение на базе абсолютной информации, равной информации у всех других агентов (то есть еще и получаемой без отставания), то тогда действительно каждый агент придет к одному и тому же решению и биржевой стакан не нужен. Но тогда и агенты не нужны — можно фиксированный план в одном месте генерировать и только на исполнение отдавать.

А в децентрализованной системе информация у любого агента неполная, вычислительные мощности каждого агента ограничены, и только система целиком как целое работает на поиск глобального оптимума (и то не абсолютного, а наилучшего из локальных в рамках временных и мощностных ограничений).

Ну и конечно предпочтения конечных потребителей — потребители тоже должны как-то сообщать, какой заказ им важнее, так как агенты не могут располагать информацией о субъективных предпочтениях. Отсюда потребительские токены, которые распределяются самими потребителями между заказами и несут для системы информацию об их относительной ценности, чтобы каждый мог сам выбрать, что ему "протолкнуть", а что получить по остаточному принципу.

## Решение

Используем агентов-закзчиков, а не единую систему распределения

# DR002: Размещение заказа

## Контекст

В случае если технологическая карта требует больше одного типа мощности для производства,то заказ может выполняться ОЧЕНЬ ДОЛГО, до тех пор пока мощности не разовьтся до уровня необходимого чтобы обработать заказа с минимальной ценой, так как одна часть будет взята в работу производителем, а оставшихся токенов не хватит чтобы его довыполнить на других мощностях в следующих тактах (так как агент-заказчик не может перераспределять токены между заказами)

## Решение

Возможные варианты:

1. агент-заказчик заключает предварительные контракты (получает ack) со всеми агентами производителями, поставщиками соответвующих мощностей и если это удалось то отдает заказ в работу. Это слишком усложняет систему, так как фазы заказа и производства начинают пересекаться
2. недовыполненный заказ имеет TTL в тактах и после этого сгорает, токены не переданные производителям возвращаются потребителю

Первое решение на самом деле более предпочтительное в целевой картине. Изначально предполагалось, что оно будет реализовано примерно так:

- агенты-производители продают через биржу свои мощности
- агенты-заказчики покупают через биржу эти мощности под свои нужды (т.е. по сути это уже и есть резерв)
- на бирже складывается какое-то ценовое равновесие
- заказ уходит в работу только если выкуплены необходимые мощности всех нужных производителей

Можно преобразовать эта систему в такой вариант:
Если заказ не попадает в работу в данный такт у ВСЕХ необходимых агентов-производителей, то он не берется никем, и тогда агенты берут следующие заказы по приоритетам. А этот уходит на переоценку и перераспределение в следующий такт. В принципе это тождественно заключению предварительных контрактов. Но тут тоже может быть, наверное, deadlock. Также это означает, что тут есть итеративный пересчет плана, который может гоняться неопределенный объем времени (так как исключение каждого заказа требует полного пересчета всех заказов у агентов, слава богу, не циклическое, так как приоритеты однонаправлены).

## Решение

Для стартовой модели используем второй вариант с TTL как временное решение.
