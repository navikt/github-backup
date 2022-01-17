from json import load as parse_json
from json.decoder import JSONDecodeError
from requests import get as httpget
from os import makedirs, environ
from os.path import exists as path_exists
from subprocess import Popen as start_process, PIPE
from sys import exit
from argparse import ArgumentParser

EXIT_SUCCESS = 0
EXIT_FAILURE = 1


def chunks(list, size):
    """Split a list into chunks"""

    for i in range(0, len(list), size):
        yield list[i:i + size]


def get_repos_in_org(org, token):
    """Get a list of repositories from a given organization

    See https://docs.github.com/en/rest/reference/repos#list-organization-repositories for the format of the dictionaries in the returned list
    """

    url = f'https://api.github.com/orgs/{org}/repos'
    query = {'per_page': 100}
    headers = {'accept': 'application/vnd.github.v3+json',
               'user-agent': 'NAV IT Backup',
               'authorization': f'bearer {token}'}
    repos = []

    while url:
        resp = httpget(url=url, params=query, headers=headers)
        repos += resp.json()
        url = resp.links.get('next', {}).get('url')
        query = {}  # no need for query on subsequent requests as the next-url includes them

    return repos


def get_repos(orgs, github_token):
    """Get all repositories

    Keyword arguments:
    orgs -- a list of dicts with name and an optional exclude list
    github_token -- a personal access token used for API authentication
    """
    repos = []

    for org in orgs:
        exclude = org.get('exclude', [])
        org_name = org.get('name')

        repos += filter(
            lambda repo: repo['name'] not in exclude,
            get_repos_in_org(org=org_name, token=github_token)
        )

    return repos


def create_backup_dirs(parent_dir, orgs):
    """Create directories used for backups

    First the parent_dir will be created, then each org in the list of orgs will get their own
    directory inside parent_dir
    """
    makedirs(name=f'{parent_dir}', exist_ok=True)

    for org in orgs:
        makedirs(name=f'{parent_dir}/{org}', exist_ok=True)


def backup(github_token, backup_dir, orgs, concurrent_processes=50):
    """Perform backup of repositories in one or more organizations on GitHub

    The backup is a simple git clone if the repository have never been backed up before, git pull otherwise
    """

    try:
        create_backup_dirs(parent_dir=backup_dir, orgs=map(
            lambda org: org['name'], orgs))
    except OSError:
        print(f'Unable to create backup directory: {backup_dir}')
        return EXIT_FAILURE

    repos = get_repos(orgs=orgs, github_token=github_token)
    num_repos = len(repos)
    counter = 1

    for chunk in chunks(list=repos, size=concurrent_processes):
        processes = []

        for repo in chunk:
            repo_name_with_owner = repo['full_name']
            repo_dir = f'{backup_dir}/{repo_name_with_owner}'

            if path_exists(repo_dir):
                command = ['git', '-C', repo_dir, 'pull', '-f']
            else:
                command = ['git', 'clone', repo['clone_url'], repo_dir]

            process = start_process(args=command, stdout=PIPE, stderr=PIPE)
            processes.append((process, repo_name_with_owner))

        for process, repo in processes:
            print(f'{counter}/{num_repos} Backing up {repo}')
            process.communicate()
            counter += 1

    return EXIT_SUCCESS


def main():
    parser = ArgumentParser(
        description='Perform a backup of repositories in GitHub organizations')
    parser.add_argument('--config-file', required=True, metavar='path', type=str,
                        help='path to configuration file')
    parser.add_argument('--backup-dir', metavar='path', type=str,
                        help='path to backup directory, defaults to /tmp/backups/github.com', default='/tmp/backups/github.com')
    parser.add_argument('--concurrent', metavar='num', type=int,
                        help='number of concurrent clones / pulls from GitHub, defaults to 50', default=50)
    args = parser.parse_args()

    if not 'GITHUB_TOKEN' in environ:
        print('Missing required environment variable GITHUB_TOKEN')
        return EXIT_FAILURE

    config_file_path = args.config_file

    try:
        config_file = open(config_file_path)
    except FileNotFoundError:
        print(f'Unable to open configuration file: {config_file_path}')
        parser.print_usage()
        return EXIT_FAILURE

    try:
        config = parse_json(config_file)
    except JSONDecodeError:
        print(
            f'Unable to parse configuration file as JSON: {config_file_path}')
        parser.print_usage()
        return EXIT_FAILURE

    return backup(github_token=environ.get('GITHUB_TOKEN'), backup_dir=args.backup_dir,
                  orgs=config['orgs'], concurrent_processes=args.concurrent)


if __name__ == '__main__':
    exit_code = main()
    exit(exit_code)
