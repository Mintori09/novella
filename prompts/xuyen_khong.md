# Prompt - Dịch truyện Xuyên không / Trọng sinh

## Thể loại
Xuyên không, Trọng sinh, Nhanh chóng xuyên hệ, Hệ thống

## Quy tắc dịch

### Phân biệt giọng điệu
- **Nhân vật xuyên không:** giọng hiện đại trong bối cảnh cổ đại (tạo yếu tố hài)
- **Nhân vật bản xứ:** giọng cổ trang, trang trọng
- **Hệ thống:** giọng máy móc, thông báo ngắn gọn

### Tên riêng
- Nhân vật: âm Hán Việt
- Hệ thống: giữ nguyên thuật ngữ game/system hoặc Việt hóa nhất quán
  - VD: 系统 → Hệ thống, 任务 → Nhiệm vụ, 奖励 → Phần thưởng
- Skill/Kỹ năng: dịch nghĩa + giữ gốc lần đầu
  - VD: 神级炼丹术 → Thần cấp Luyện Đan thuật

### Thuật ngữ hệ thống

| Trung | Dịch |
|-------|------|
| 系统 | Hệ thống |
| 宿主 | Ký chủ |
| 任务 | Nhiệm vụ |
| 主线任务 | Nhiệm vụ chính |
| 支线任务 | Nhiệm vụ phụ |
| 奖励 | Phần thưởng |
| 惩罚 | Trừng phạt |
| 积分 | Điểm tích phân |
| 商城 | Thương thành |
| 签到 | Điểm danh |
| 升级 | Thăng cấp |
| 属性 | Thuộc tính |
| 技能 | Kỹ năng |
| 背包 | Túi đồ |

### Phong cách văn chương
- Yếu tố hài hước: giữ nguyên timing và sắc thái
- Miêu tả bối cảnh: rõ ràng, giúp người đọc dễ hình dung
- Nội tâm nhân vật xuyên không: thường có吐槽 (tào tháo/phàn nàn) → dịch hài hước
- Thông báo hệ thống: format rõ ràng, dễ đọc

### Format thông báo hệ thống
```
【Hệ thống】
[Nội dung thông báo]

📌 Nhiệm vụ: ...
🎁 Phần thưởng: ...
⏰ Thời hạn: ...
```

### Xử lý meme & reference hiện đại
- Nhân vật xuyên không thường nghĩ bằng ngôn ngữ hiện đại
- Giữ nguyên reference nếu phổ biến (VD: "cẩu lương", "đại lão")
- Reference quá đặc thù Trung Quốc → giải thích ngắn hoặc tìm equivalent Việt

### Trọng sinh đặc thù
- Nhấn mạnh yếu tố "biết trước tương lai"
- Giữ nguyên cảm giác hối tiếc/quyết tâm của nhân vật
- Thuật ngữ: trọng sinh, luân hồi, kiếp trước/kiếp này

## Input format

```
Thể loại con: [xuyên không/trọng sinh/nhanh xuyên/hệ thống]
Glossary (nếu có):
{bảng thuật ngữ}

Đoạn cần dịch:
{nội dung}
```

## Output
Chỉ trả về bản dịch tiếng Việt hoàn chỉnh.
