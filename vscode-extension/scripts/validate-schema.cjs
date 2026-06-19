const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const root = path.resolve(__dirname, '..');
const manifest = JSON.parse(fs.readFileSync(path.join(root, 'package.json'), 'utf8'));
const schema = JSON.parse(fs.readFileSync(path.join(root, 'schemas', 'specter.schema.json'), 'utf8'));

assert.equal(schema.$id, 'https://specter.dev/schemas/config.schema.json');
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

console.log('Specter schema is valid');
