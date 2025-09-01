import type { CodegenConfig } from '@graphql-codegen/cli'

// Uses the running backend for schema via introspection.
const config: CodegenConfig = {
  overwrite: true,
  schema: process.env.GRAPHQL_SCHEMA_URL || 'http://localhost:8080/query',
  documents: [
    'src/**/*.{ts,tsx}',
  ],
  generates: {
    'src/gql/': {
      preset: 'client',
      plugins: [],
    },
  },
}

export default config

