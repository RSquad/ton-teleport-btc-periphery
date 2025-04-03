#include <stdarg.h>
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>


typedef struct Buffer {
  const uint8_t *data;
  size_t len;
} Buffer;

typedef struct Pkg {
  uint8_t identifier[32];
  const uint8_t *buf;
  size_t len;
} Pkg;

int32_t aggregate_with_tweak(struct Buffer message_buf,
                             const struct Pkg *commitments_ptr,
                             size_t commitments_len,
                             const struct Pkg *signature_shares_ptr,
                             size_t signature_shares_len,
                             struct Buffer pubkey_package_buf,
                             struct Buffer merkle_root_buf,
                             struct Buffer *signature_buf,
                             uint8_t (*culprit_idx_out)[32]);

int32_t commit(struct Buffer key_package_buf,
               struct Buffer *nonces_buf,
               struct Buffer *commitments_buf);

int32_t dkg_part1(const uint8_t (*identifier)[32],
                  uint16_t min_signers,
                  uint16_t max_signers,
                  uint8_t *package,
                  int32_t package_len,
                  const void **secret_package);

int32_t dkg_part2(const void *r1_secret,
                  const struct Pkg *r1_pkgs_ptr,
                  size_t r1_pkgs_len,
                  const struct Pkg **r2_pkgs_ptr,
                  const void **r2_secret,
                  uint8_t (*r2_culprit_idx_out)[32]);

int32_t dkg_part3(const void *r2_secret,
                  const struct Pkg *r1_pkgs_ptr,
                  size_t r1_pkgs_len,
                  const struct Pkg *r2_pkgs_ptr,
                  size_t r2_pkgs_len,
                  const uint8_t **public_key_pkg_ptr,
                  size_t *public_key_pkg_len,
                  const uint8_t **secret_key_pkg_ptr,
                  size_t *secret_key_pkg_len,
                  uint8_t (*r3_culprit_idx_out)[32]);

int32_t ext_get_identifier(uint16_t key, uint8_t (*identifier)[32]);

int32_t extract_public_key_from_package(struct Buffer pubkey_package_buf,
                                        struct Buffer *public_key);

void free_package_ptr(const uint8_t *ptr, size_t len);

void free_r2_pkg_vec(const struct Pkg *ptr, size_t len);

void free_r2_secret(void *r2_secret);

int32_t sign_with_tweak(struct Buffer key_package_buf,
                        struct Buffer message_buf,
                        const struct Pkg *commitments_ptr,
                        size_t commitments_len,
                        struct Buffer nonces_buf,
                        struct Buffer merkle_root_buf,
                        struct Buffer *signature_share_buf);

int32_t verify(struct Buffer message_buf,
               struct Buffer pubkey_package_buf,
               struct Buffer signature_buf,
               struct Buffer merkle_root_buf);
