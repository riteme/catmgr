#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import json
import datetime
import textwrap

from typing import *

import click
import requests


VERBOSE = False
CONFIG_FILE = 'catmgr.json'
SERVER_URL = 'http://localhost:10777/'
DEFAULT_USER = None
DEFAULT_PASSWORD = None

if os.path.exists(CONFIG_FILE):
    with open(CONFIG_FILE, 'r') as fp:
        config = json.load(fp)

    if 'server_url' in config:
        SERVER_URL = config['server_url']
    if 'user' in config:
        DEFAULT_USER = config['user']
    if 'password' in config:
        DEFAULT_PASSWORD = config['password']


JSONMap = Mapping[str, Any]

def invoke(route: str, payload: JSONMap) -> JSONMap:
    if VERBOSE:
        print(f'→ "{route}", Payload: {payload}')

    r = requests.post(os.path.join(SERVER_URL, route), json=payload)
    r.raise_for_status()
    return r.json()

def print_error(resp: JSONMap) -> None:
    print(f'Status: {resp["status"]}')
    print(f'Reason: {resp["error"]}')

def print_book(book: JSONMap) -> None:
    print(textwrap.dedent(f'''
        #{book["book_id"]}
        Title:\t{book["title"]}
        Author:\t{book["author"]}
        ISBN:\t{book["isbn"]}
        Count:\t{book["count"]}
        Comment: {book["comment"]}
        Description:
        {book["description"]}'''))

def parse_date(val: str) -> datetime.date:
    date_val = val.split('T', maxsplit=1)[0]
    parsed = map(int, date_val.split('-'))
    return datetime.date(*parsed)

def print_record(record: JSONMap) -> None:
    returned = record['returned']
    if returned:
        return_date = parse_date(record['return_date'])

    borrow_date = parse_date(record['borrow_date'])
    due_date = parse_date(record['deadline'])
    today = datetime.date.today()

    if returned:
        status = 'returned'
    elif due_date < today:
        status = 'overdue'
    else:
        status = 'normal'

    book_id = record['book_id']

    print(textwrap.dedent(f'''
        #{record["record_id"]}
        User:\t{record["username"]}
        Status:\t{status}
        Borrow:\t{borrow_date}
        Due:\t{due_date}'''))
    if returned:
        print(f'Return:\t{return_date}')

    print(f'Book ID: #{book_id}')
    try:
        resp = invoke('show', {
            'section': 'book_id',
            'keyword': str(book_id)
        })

        if resp['status'] == 'ok' and len(resp['results']) == 1:
            book = resp['results'][0]
            print(f'Title:\t{book["title"]}')
            print(f'Author:\t{book["author"]}')
        else:
            raise Exception
    except:
        print(f'(failed to retrieve book #{book_id})')


user_prompt = click.option('-u', '--user', type=str, default=DEFAULT_USER, show_default='set in "catmgr.json"', prompt=True,
    help='Account username.')
password_prompt = click.option('-p', '--password', type=str, default=DEFAULT_PASSWORD, show_default='set in "catmgr.json"', prompt=True, hide_input=True,
    help='Account password.')

@click.group()
@click.option('-v', '--verbose', is_flag=True,
    help='Show more information.')
def cli(verbose: bool) -> None:
    global VERBOSE
    VERBOSE = verbose

@cli.command(short_help='Add a new book.')
@user_prompt
@password_prompt
def new(**kwargs) -> None:
    resp = invoke('new', kwargs)

    if resp['status'] == 'ok':
        print(f'New book: #{resp["book_id"]}')
    else:
        print_error(resp)

@cli.command(short_help='Update book information.')
@user_prompt
@password_prompt
@click.argument('book_id', type=int)
@click.option('--diff', type=int,
    help='Difference of available number of the book.')
@click.option('--title', type=str,
    help='Book title.')
@click.option('--author', type=str,
    help='Author of the book.')
@click.option('--isbn', type=str,
    help='ISBN of the book.')
@click.option('--description', '--desc', type=str,
    help='Description/overview of the book.')
@click.option('--comment', type=str,
    help='Comment of the book.')
def update(**kwargs) -> None:
    resp = invoke('update', kwargs)

    if resp['status'] == 'ok':
        print(f'Update book: #{resp["book_id"]}')
    else:
        print_error(resp)

@cli.command(short_help='Add a new user.')
@user_prompt
@password_prompt
@click.option('--new-user-type', prompt=True, type=str,
    help='User type of the new user.')
@click.option('--new-username', prompt=True, type=str,
    help='Username for the new user.')
@click.option('--new-password', prompt=True, hide_input=True, confirmation_prompt=True, type=str,
    help='Password for the new user.')
def adduser(**kwargs) -> None:
    resp = invoke('adduser', kwargs)

    if resp['status'] == 'ok':
        print(f'New user: "{kwargs["new_username"]}" #{resp["user_id"]}')
    else:
        print_error(resp)

@cli.command(short_help='Search for books.')
@click.option('-s', '--section', type=click.Choice(['book_id', 'isbn', 'title', 'author']), required=True,
    help='Section to be searched.')
@click.argument('keyword', type=str)
def show(**kwargs) -> None:
    resp = invoke('show', kwargs)

    if resp['status'] != 'ok':
        print_error(resp)
        return

    results = resp['results']
    for book in results:
        print_book(book)
    print(f'\n{len(results)} result(s)')

@cli.command(name='list', short_help='List borrow history.')
@user_prompt
@password_prompt
@click.argument('target', type=str)
@click.option('-f', '--filter', type=click.Choice(['all', 'not-returned', 'overdue']), default='all', show_default=True,
    help='Filter condition.')
@click.option('-l', '--limit', type=click.IntRange(min=0), default=100, show_default=True,
    help='Maximum number of records to be returned by the server.')
def list_(**kwargs) -> None:
    resp = invoke('list', kwargs)

    if resp['status'] != 'ok':
        print_error(resp)
        return

    results = resp['results']
    for record in results:
        print_record(record)
    print(f'\n{len(results)} result(s)')

@cli.command(short_help='Borrow a book.')
@user_prompt
@password_prompt
@click.argument('book_id', type=int)
def borrow(**kwargs) -> None:
    resp = invoke('borrow', kwargs)

    if resp['status'] == 'ok':
        print(f'Success! Record ID: #{resp["record_id"]}')
    else:
        print_error(resp)

@cli.command(short_help='Extend deadline.')
@user_prompt
@password_prompt
@click.argument('record_id', type=int)
def extend(**kwargs) -> None:
    resp = invoke('extend', kwargs)

    if resp['status'] == 'ok':
        print(f'Record deadline extended: #{resp["record_id"]}')
    else:
        print_error(resp)

@cli.command(name='return', short_help='Return a book.')
@user_prompt
@password_prompt
@click.argument('record_id', type=int)
def return_(**kwargs) -> None:
    resp = invoke('return', kwargs)

    if resp['status'] == 'ok':
        print(f'Book returned. Record ID: #{resp["record_id"]}')
    else:
        print_error(resp)


if __name__ == '__main__':
    cli()
