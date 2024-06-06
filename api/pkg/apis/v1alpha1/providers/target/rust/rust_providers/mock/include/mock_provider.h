#ifndef MOCK_PROVIDER_H
#define MOCK_PROVIDER_H

#ifdef __cplusplus
extern "C" {
#endif

void* create_mock_provider();
void destroy_mock_provider(void* provider);

#ifdef __cplusplus
}
#endif

#endif // MOCK_PROVIDER_H
