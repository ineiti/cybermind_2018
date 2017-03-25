# CyberMind

CyberMind's goal is to make any data coming into your computer easily
aggregatable. Data-sources are text-notes, emails, photos, voice,
Whatsapp, Signal, ... Every source stays with its own creator but can
be linked with other sources through tags.

Tags are added automatically (for date, place, source), semi-automatically
(for detected keywords) or manually.

Roadmap 2017:
- January: write down ideas, start backend
- February: implement simple web-client
- March: make usable Android / iOs client
- April: make desktop-app for Windows/Mac/Linux
- Mai: add email plugin
- June: add bookmark/www plugin

# Architecture

## Sources

Different sources can be added as plug-ins to cybermind:

- text - most simple source where the user types in text
- email - probably on server - fetches email from POP or IMAP and
 adds new emails to the database.
- bookmarks - choosing a page and/or select text

They must have some of the following:

- Input - user-input or mail-server or browser-plugin

## Backend

Database storage for all sources. Some ideas:

- make it directories/files for easy synchronisation across clouds
- every device syncs its files between locally and the cloud
- after synching, the device checks for conflicts
- file-name is either $( date )_${version}_${device} or a hash of
  the file
- there could be two files: one .json and one .data in case the
  corresponding data is bigger than 1k

## Controller

The controller plugs the different sources together and handles the
tags.

## Frontend

The frontend only displays the

## Tags


### Automatic

A couple of automatic tags which are added always - can they be deleted?

- date/time
- device-name where the document has been created
- service-type and service-name

### Semi-automatic

Semi-automatic tags, presented to the user - perhaps for 'sure' tags they
are added by default, for 'not sure' tags, the user is asked.

- rare words present in the document
- tags present in similar documents

### Manual tags

All other, can be added at wish.

### Tag-hierarchy

Tags can depend on other tags - when the child-tag is added, the parent-
tag is also present.

## Filters and actions

The user can define filters that do actions on sources:

### Actions

- email
    - move to folder
    - delete
    - mark as SPAM
    - auto-reply
- IFTTT
    - send actions to IFTTT
    
### Filters

Do standard-filters on content (regexp), date, ...
