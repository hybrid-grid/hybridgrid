#!/bin/bash
# Record or verify analyzer commands and exact outputs of required families.
set -u
if (( $# < 3 )) || [[ $1 != write && $1 != check ]]; then
    echo 'usage: analysis_marker.sh write|check <out_dir> <required_subdirs...> --argv <family=arguments...>' >&2
    exit 2
fi
mode=$1; out=$2; shift 2
: "${BENCH_REPO:?BENCH_REPO is required}"
python3 - "$mode" "$out" "$BENCH_REPO/scripts/analyze_exp.py" "$@" <<'PY'
import hashlib, json, os, sys, tempfile

mode, out, analyzer, *args = sys.argv[1:]
if '--argv' not in args:
    print('analysis marker: missing --argv', file=sys.stderr)
    sys.exit(2)
sep = args.index('--argv')
families, argv_items = args[:sep], args[sep + 1:]
if not families or len(argv_items) != len(families):
    print('analysis marker: family/argv count mismatch', file=sys.stderr)
    sys.exit(2)
argv = {}
for item in argv_items:
    family, delim, value = item.partition('=')
    if not delim or family not in families or family in argv or not value:
        print('analysis marker: invalid family argv', file=sys.stderr)
        sys.exit(2)
    argv[family] = ' '.join(value.split())
marker = os.path.join(out, 'analysis.done')

def digest(path):
    h = hashlib.sha256()
    with open(path, 'rb') as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b''):
            h.update(chunk)
    return h.hexdigest()

def snapshot():
    if not os.path.isfile(analyzer):
        raise ValueError('analyzer missing')
    files = {}
    for family in families:
        if family in ('', '.', '..') or '/' in family or '\\' in family:
            raise ValueError('invalid family')
        directory = os.path.join(out, family)
        log = os.path.join(out, family + '.log')
        if not os.path.isdir(directory) or not os.path.isfile(log):
            raise ValueError('family or log missing')
        files[family + '.log'] = digest(log)
        count = 0
        for root, dirs, names in os.walk(directory):
            dirs.sort()
            names.sort()
            if any(os.path.islink(os.path.join(root, name)) for name in dirs):
                raise ValueError('symlinked analysis directory')
            for name in names:
                path = os.path.join(root, name)
                if os.path.islink(path) or not os.path.isfile(path):
                    raise ValueError('non-file output')
                files[os.path.relpath(path, out).replace(os.sep, '/')] = digest(path)
                count += 1
        if count == 0:
            raise ValueError('empty family')
    return {'analyzer_sha256': digest(analyzer), 'families': families, 'argv': argv, 'files': files}

try:
    actual = snapshot()
    if mode == 'check':
        with open(marker, encoding='utf-8') as f:
            expected = json.load(f)
        if expected != actual:
            raise ValueError('analysis marker mismatch')
    else:
        fd, tmp = tempfile.mkstemp(prefix='.analysis.done.', dir=out)
        try:
            with os.fdopen(fd, 'w', encoding='utf-8') as f:
                json.dump(actual, f, sort_keys=True, indent=2)
                f.write('\n')
            os.replace(tmp, marker)
        finally:
            if os.path.exists(tmp):
                os.unlink(tmp)
except (OSError, ValueError, json.JSONDecodeError) as exc:
    print(f'analysis marker: {exc}', file=sys.stderr)
    sys.exit(1)
PY
