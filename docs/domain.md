# Domain

“Think like a language teacher”

## Main concepts

User

- email
- password
- name

Deck

- Cards
- visibility
- name
- description

Card (abstract) → specialization: VocabularyCard, KanjiCard, GrammarCard

- question
- answer
- Type
- tags

Vocabulary

- pronounciation/pitch
- example sentence
- JLPT level
- tags

GrammarPoint

- example sentence
- patterns
- JLPT level

Kanji

- mnemonic
- furigana
- example sentence

StudySession

- start
- end
- accuracy
- cards reviewd
- duration

Review

- Card
- result
- Time spent
- Reviewed at
- next review
- state

Progress

- goal note to motivate
- point
- grammar learned
- kanji learned
- days past
- current percentage reached

Report

- Reviews
- Progress

### Initial Database Model

Entities: users, decks, cards, reviews, study_sessions, reports. Next, how do they relate to each other?

## Event Storming

> These become: domain events, audit logs, notifications later
> 

UserRegistered

LoggedIn

DeckCreated

CardAdded

StudyStarted

StudyFinished

CardReviewed

ReportGenerated