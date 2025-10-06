import { defineConfig } from 'orval';

export default defineConfig({
  nexusAPI: {
    input: '../../docs/openapi.yaml', 
    output: {
      mode: 'tags-split',
      target: './src/api/endpoints/api.ts',
      schemas: './src/api/model',
      client: 'react-query',
      httpClient: 'axios',
      prettier: true,
      clean: true,
      override: {
        mutator: {
          path: './src/lib/axios.ts',
          name: 'customInstance'
        }
      }
    }
  }
});
