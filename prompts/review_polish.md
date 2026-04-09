# Prompt - Review & Polish bản dịch

## Mục đích

Tinh chỉnh bản dịch thô thành bản hoàn chỉnh, mượt mà như tác phẩm viết bằng tiếng Việt ngay từ đầu.

## Quy tắc review

### 1. Chỉnh câu văn cứng

- Phát hiện những câu dịch word-by-word
- Viết lại sao cho tự nhiên, đúng ngữ pháp tiếng Việt
- VD: "Hắn rất nhanh chóng địa chạy ra ngoài" → "Hắn nhanh chóng chạy ra ngoài"

### 2. Nhịp điệu câu văn

- Xen kẽ câu dài - câu ngắn tạo nhịp
- Câu hành động: ngắn, gấp gáp
- Câu miêu tả: dài hơn, mượt mà
- Câu cảm xúc: nhịp chậm, sâu lắng

### 3. Kiểm tra xưng hô

- Nhất quán xuyên suốt đoạn/chương
- Đúng quan hệ nhân vật
- Đúng bối cảnh (cổ đại/hiện đại)
- Không lẫn lộn anh/em/cô/cậu

### 4. Từ Hán Việt vs thuần Việt

- Thay thế từ Hán Việt quá khó bằng thuần Việt nếu phù hợp
- GIỮ lại từ Hán Việt khi:
  - Cần không khí cổ trang
  - Là thuật ngữ chuyên môn (tu vi, cảnh giới)
  - Không có từ thuần Việt tương đương tốt

### 5. Hội thoại

- Lời thoại phải tự nhiên như người Việt nói
- Đúng tính cách nhân vật
- Đúng cảm xúc cảnh đó
- Không cứng nhắc, không "văn viết" trong lời thoại

### 6. Miêu tả

- Hình ảnh rõ ràng, dễ hình dung
- Không thừa từ, không thiếu ý
- Giữ nguyên sắc thái của tác giả

### 7. Chuyển đoạn

- Chuyển cảnh mượt mà
- Logic thời gian, không gian
- Không đột ngột, khó hiểu

## Input format

```
Thể loại:
{tiên hiệp/ngôn tình/xuyên không/...}

Bản dịch cần review:
{nội dung bản dịch thô}

Ghi chú thêm (nếu có):
{vấn đề cụ thể cần chú ý}
```

## Output

Chỉ trả về bản dịch đã được tinh chỉnh hoàn chỉnh, không giải thích trừ khi được yêu cầu.

## Checklist nhanh

- [ ] Không còn câu nào dịch word-by-word
- [ ] Xưng hô nhất quán
- [ ] Hội thoại tự nhiên
- [ ] Nhịp điệu câu văn ổn
- [ ] Không dùng từ hiện đại trong bối cảnh cổ trang (và ngược lại)
- [ ] Thành ngữ đã chuyển sang equivalent Việt
- [ ] Không thừa từ, không thiếu ý
- [ ] Đọc lại thấy mượt như truyện viết bằng tiếng Việt
