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


def reorder_properties(props: dict, order: list, inject_x_order: bool = False) -> dict:
    """Return props dict with keys re-ordered according to order list.

    inject_x_order: 同时在每个 property 上写入 x-order，让 Scalar/Swagger UI
    等按字母序渲染的工具也能遵循正确顺序。
    """
    ordered = {k: props[k] for k in order if k in props}
    ordered.update({k: v for k, v in props.items() if k not in order})
    if inject_x_order:
        for i, key in enumerate(ordered):
            ordered[key]['x-order'] = i
    return ordered


def flatten_success_allof(schema: dict, definitions: dict, order: list) -> dict:
    """将引用 APISuccessResponse 的 allOf 结构展开为单一 flat object。

    Swagger UI 对 allOf 合并后的字段展示顺序由其自身逻辑决定（通常字母序），
    无法从外部干预。展开为 flat object 后，字段顺序完全由 JSON key 顺序保证。
    只处理 allOf = [$ref, {properties}] 这种 swag 生成的固定模式。
    """
    all_of = schema.get('allOf', [])
    if len(all_of) < 2:
        return schema

    # 收集 $ref 指向的 definition properties
    merged_props = {}
    for sub in all_of:
        if '$ref' in sub:
            ref_name = sub['$ref'].split('/')[-1]
            ref_def = definitions.get(ref_name, {})
            merged_props.update(ref_def.get('properties', {}))
        elif 'properties' in sub:
            # inline object 的字段覆盖 $ref 中的同名字段
            merged_props.update(sub['properties'])

    if not merged_props:
        return schema

    return {
        'type': 'object',
        'properties': reorder_properties(merged_props, order, inject_x_order=True),
    }


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

    # Reorder Response properties — Swagger 2.0（同时注入 x-order 供 Scalar 等工具使用）
    if 'definitions' in doc:
        for name in ['APIResponse', 'APISuccessResponse', 'APISuccessNoDataResponse']:
            if name in doc['definitions']:
                props = doc['definitions'][name].get('properties', {})
                if props:
                    doc['definitions'][name]['properties'] = reorder_properties(
                        props, RESPONSE_FIELD_ORDER, inject_x_order=True
                    )

    # Reorder Response properties — OpenAPI 3.x
    schemas = doc.get('components', {}).get('schemas', {})
    for name in ['APIResponse', 'APISuccessResponse', 'APISuccessNoDataResponse']:
        if name in schemas:
            props = schemas[name].get('properties', {})
            if props:
                schemas[name]['properties'] = reorder_properties(
                    props, RESPONSE_FIELD_ORDER, inject_x_order=True
                )

    definitions = doc.get('definitions', {})

    # 遍历所有 path 的 200 response，将 allOf 展开为 flat object
    # 原因：swag 为带具体 data 的成功响应生成 allOf 结构，Swagger UI 合并字段时
    # 按自身逻辑排序（通常字母序），无法通过 definitions 里的顺序控制。
    # 展开后字段顺序完全由 JSON key 顺序决定，确保信封字段 success→code→message→data。
    for methods in doc.get('paths', {}).values():
        for op in methods.values():
            schema = op.get('responses', {}).get('200', {}).get('schema', {})
            if schema.get('allOf'):
                op['responses']['200']['schema'] = flatten_success_allof(
                    schema, definitions, RESPONSE_FIELD_ORDER
                )

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
