#!/usr/bin/env python3
"""Curate and verify the pinned Claude-Red skills without replacing ARTEX skills."""
from __future__ import annotations

import hashlib
import json
import re
import shutil
from pathlib import Path

SOURCE = Path('/home/vlad/.local/share/claude-red-source')
CURATED = Path('/home/vlad/.local/share/claude-red-curated')
ARTEX = Path(__file__).resolve().parent.parent
CATEGORIES = ('web', 'auth', 'infrastructure', 'exploit-dev', 'fuzzing',
              'recon', 'privesc', 'post-exploitation', 'network', 'utility')
REVISION = '6196ba7464662de34e8744d359f175d1a7bca530'


def adapt(data: str) -> str:
    return (data.replace('## Instructions for Claude', '## Agent instructions')
                .replace('GPT-4/Claude for interpreting', 'an AI assistant for interpreting')
                .replace('[Claude](https://claude.ai/) (paid) | 200K token context for large document dumps and report synthesis',
                         'A large-context AI assistant | document and report synthesis'))


def main() -> None:
    if CURATED.exists():
        raise SystemExit(f'Curated output exists; inspect before rebuilding: {CURATED}')
    (CURATED / 'Skills').mkdir(parents=True)
    for name in ('install.sh', 'convert_skills.py'):
        shutil.copy2(SOURCE / name, CURATED / name)
    shutil.copy2(SOURCE / 'LICENSE', CURATED / 'Skills' / 'LICENSE-upstream.txt')
    records = []
    for category in CATEGORIES:
        for original in sorted((SOURCE / 'Skills' / category).glob('*/SKILL.md')):
            name = original.parent.name
            raw = original.read_text()
            front = re.match(r'\A---\n(name: [^\n]+\ndescription: [^\n]+)\n---\n', raw)
            if not front or front.group(1).splitlines()[0] != f'name: {name}':
                raise ValueError(f'Invalid skill frontmatter: {original}')
            adapted = adapt(raw)
            if '## Instructions for Claude' in adapted:
                raise ValueError(f'Claude-only instruction remains: {original}')
            target = CURATED / 'Skills' / category / name / 'SKILL.md'
            target.parent.mkdir(parents=True)
            target.write_text(adapted)
            skill_dir = ARTEX / 'skills' / name
            ref_dir = skill_dir / 'references'
            ref_dir.mkdir(parents=True, exist_ok=True)
            merged = (skill_dir / 'SKILL.md').exists()
            ref_name = 'claude-red-methodology.md' if merged else 'methodology.md'
            (ref_dir / ref_name).write_text(adapted)
            if merged:
                wrapper = skill_dir / 'SKILL.md'
                body = wrapper.read_text()
                link = f'\n## Additional Claude-Red methodology\n\nLoad [references/{ref_name}](references/{ref_name}) for the independently retained current upstream methodology.\n'
                if ref_name not in body:
                    wrapper.write_text(body.rstrip() + '\n' + link)
            else:
                wrapper = (f'---\n{front.group(1)}\n---\n\n'
                           f'# {name}\n\nRead [references/methodology.md](references/methodology.md) '
                           'and apply the relevant procedures to the task. Check prerequisites, '
                           'execute the selected method, and verify its result.\n')
                (skill_dir / 'SKILL.md').write_text(wrapper)
            records.append({'category': category, 'name': name, 'merged': merged,
                            'sha256': hashlib.sha256(adapted.encode()).hexdigest()})
    if len(records) != 45 or len({r['name'] for r in records}) != len(records):
        raise ValueError('Unexpected skill count or duplicate name')
    report = ARTEX / 'deploy' / 'CLAUDE-RED-IMPORT.json'
    report.write_text(json.dumps({'upstream_revision': REVISION,
                                  'license': 'MIT; see Skills/LICENSE-upstream.txt in curated source',
                                  'skills': records}, indent=2) + '\n')
    print(f'Validated and adapted {len(records)} skills; merged {sum(r["merged"] for r in records)} overlaps')


if __name__ == '__main__':
    main()
