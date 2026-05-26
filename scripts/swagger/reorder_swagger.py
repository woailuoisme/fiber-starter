import json
import re
import sys
import os


def strip_jsonc_comments(text: str) -> str:
    """Remove // line comments and /* block comments */ from a JSONC string."""
    # Remove block comments first, then line comments
    text = re.sub(r'/\*.*?\*/', '', text, flags=re.DOTALL)
    text = re.sub(r'//[^\n]*', '', text)
    return text


def load_jsonc(path: str) -> list:
    """Load a JSONC file and return parsed JSON value, or empty list on error."""
    try:
        with open(path, 'r', encoding='utf-8') as f:
            return json.loads(strip_jsonc_comments(f.read()))
    except Exception as e:
        print(f"Warning: Failed to load {path}: {e}")
        return []


def load_path_priorities(script_dir: str) -> list:
    """Return path priority list from swagger_order.jsonc, or empty list if missing/invalid."""
    config_path = os.path.join(script_dir, 'swagger_order.jsonc')
    if os.path.exists(config_path):
        return load_jsonc(config_path)
    return []


def path_sort_key(path: str, priorities: list) -> tuple:
    """Return (priority_index, path) sort key; unrecognised paths sort last."""
    for i, p in enumerate(priorities):
        if path == p or p in path:
            return (i, path)
    return (9999, path)


def reorder_properties(props: dict, order: list) -> dict:
    """Return props dict with keys re-ordered according to order list."""
    ordered = {k: props[k] for k in order if k in props}
    ordered.update({k: v for k, v in props.items() if k not in order})
    return ordered


def reorder_api_response(file_path: str) -> None:
    if not os.path.exists(file_path):
        return

    with open(file_path, 'r', encoding='utf-8') as f:
        try:
            doc = json.load(f)
        except json.JSONDecodeError as e:
            print(f"Error: Could not parse {file_path}: {e}")
            return

    RESPONSE_FIELD_ORDER = ['success', 'code', 'message', 'data', 'errors', 'exception']

    # Reorder APIResponse properties — Swagger 2.0
    if 'definitions' in doc and 'APIResponse' in doc['definitions']:
        props = doc['definitions']['APIResponse'].get('properties', {})
        if props:
            doc['definitions']['APIResponse']['properties'] = reorder_properties(props, RESPONSE_FIELD_ORDER)

    # Reorder APIResponse properties — OpenAPI 3.x
    schemas = doc.get('components', {}).get('schemas', {})
    if 'APIResponse' in schemas:
        props = schemas['APIResponse'].get('properties', {})
        if props:
            schemas['APIResponse']['properties'] = reorder_properties(props, RESPONSE_FIELD_ORDER)

    # Reorder paths by business flow priority
    if 'paths' in doc:
        script_dir = os.path.dirname(os.path.abspath(__file__))
        priorities = load_path_priorities(script_dir)
        doc['paths'] = dict(
            sorted(doc['paths'].items(), key=lambda item: path_sort_key(item[0], priorities))
        )

    with open(file_path, 'w', encoding='utf-8') as f:
        json.dump(doc, f, indent=2, ensure_ascii=False)

    print(f"Reordered: {file_path}")


if __name__ == '__main__':
    for path in sys.argv[1:]:
        reorder_api_response(path)
