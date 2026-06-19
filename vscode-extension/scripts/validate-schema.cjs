const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const root = path.resolve(__dirname, '..');
const repoRoot = path.resolve(root, '..');
const manifest = JSON.parse(fs.readFileSync(path.join(root, 'package.json'), 'utf8'));
const extensionSchemaPath = path.join(root, 'schemas', 'specter.schema.json');
const canonicalSchemaPath = path.join(repoRoot, 'schemas', 'config.schema.json');
const siteSchemaPath = path.join(repoRoot, 'site', 'static', 'schemas', 'config.schema.json');
const extensionSchema = fs.readFileSync(extensionSchemaPath, 'utf8');
const canonicalSchema = fs.readFileSync(canonicalSchemaPath, 'utf8');
const siteSchema = fs.readFileSync(siteSchemaPath, 'utf8');
const schema = JSON.parse(extensionSchema);

assert.equal(schema.$id, 'https://specter.dev/schemas/config.schema.json');
assert.equal(schema['x-specter-schema-version'], '1');
assert.equal(schema.type, 'object');
assert.ok(schema.properties.routes);
assert.equal(schema.$defs.route.properties.method.$ref, '#/$defs/httpMethod');
assert.ok(schema.$defs.httpMethod.enum.includes('GET'));
assert.equal(schema.$defs.route.properties.latency_profile.$ref, '#/$defs/latencyProfileName');
assert.ok(schema.$defs.latencyProfileName.anyOf[0].enum.includes('mobile-4g'));
assert.equal(schema.$defs.route.properties.fault.$ref, '#/$defs/fault');
assert.ok(schema.$defs.fault.enum.includes('timeout'));
assert.ok(schema.$defs.routeMatch.properties.body_path.additionalProperties.type === 'string');

const validations = manifest.contributes.yamlValidation;
assert.ok(Array.isArray(validations));
for (const fileMatch of ['specter.yaml', 'specter.yml', 'config.yaml', 'config.yml']) {
	assert.ok(validations.some((entry) => entry.fileMatch === fileMatch), `${fileMatch} is associated`);
}

assert.equal(extensionSchema, canonicalSchema, 'VS Code extension schema must match schemas/config.schema.json');
assert.equal(siteSchema, canonicalSchema, 'site/static schema must match schemas/config.schema.json');

console.log('Specter schema is valid');
