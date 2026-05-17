import json
import sys
import os

def reorder_api_response(file_path):
    if not os.path.exists(file_path):
        return

    with open(file_path, 'r', encoding='utf-8') as f:
        try:
            d = json.load(f)
        except json.JSONDecodeError:
            return

    # 定义目标顺序
    order = ['success', 'code', 'message', 'data', 'errors', 'exception']
    
    # 支持 Swagger 2.0 (definitions) 和 OpenAPI 3.x (components/schemas)
    found = False
    
    # 处理 Swagger 2.0
    if 'definitions' in d and 'APIResponse' in d['definitions']:
        props = d['definitions']['APIResponse'].get('properties', {})
        if props:
            ordered_props = {k: props[k] for k in order if k in props}
            ordered_props.update({k: v for k, v in props.items() if k not in order})
            d['definitions']['APIResponse']['properties'] = ordered_props
            found = True

    # 处理 OpenAPI 3.x
    if 'components' in d and 'schemas' in d['components'] and 'APIResponse' in d['components']['schemas']:
        props = d['components']['schemas']['APIResponse'].get('properties', {})
        if props:
            ordered_props = {k: props[k] for k in order if k in props}
            ordered_props.update({k: v for k, v in props.items() if k not in order})
            d['components']['schemas']['APIResponse']['properties'] = ordered_props
            found = True

    if found:
        with open(file_path, 'w', encoding='utf-8') as f:
            json.dump(d, f, indent=2, ensure_ascii=False)
        print(f"Successfully reordered APIResponse in {file_path}")

if __name__ == "__main__":
    for path in sys.argv[1:]:
        reorder_api_response(path)
