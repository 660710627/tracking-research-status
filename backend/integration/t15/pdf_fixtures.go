// PDF fixture builders copied from the existing T-06 tests; no test files modified.
package main

import (
	"bytes"
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

func pdfFixture(pages int, padding int, password *string) []byte {
	objects := []string{"<< /Type /Catalog /Pages 2 0 R >>", "<< /Type /Pages /Kids [] /Count 0 >>"}
	if pages > 0 {
		objects[1] = "<< /Type /Pages /Kids [3 0 R] /Count 1 >>"
		objects = append(objects, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>", "<< /Length 0 >>\nstream\n\nendstream")
	}
	encryptRef := ""
	id := bytes.Repeat([]byte{0x17}, 16)
	if password != nil {
		pad := []byte{0x28, 0xbf, 0x4e, 0x5e, 0x4e, 0x75, 0x8a, 0x41, 0x64, 0x00, 0x4e, 0x56, 0xff, 0xfa, 0x01, 0x08, 0x2e, 0x2e, 0x00, 0xb6, 0xd0, 0x68, 0x3e, 0x80, 0x2f, 0x0c, 0xa9, 0xfe, 0x64, 0x53, 0x69, 0x7a}
		padded := func(p string) []byte { b := append([]byte(p), pad...); return b[:32] }
		ownerHash := md5.Sum(padded("owner-password"))
		cipher, _ := rc4.NewCipher(ownerHash[:5])
		owner := make([]byte, 32)
		cipher.XORKeyStream(owner, padded(*password))
		permissions := make([]byte, 4)
		binary.LittleEndian.PutUint32(permissions, 0xfffffffc)
		material := append(padded(*password), owner...)
		material = append(material, permissions...)
		material = append(material, id...)
		key := md5.Sum(material)
		cipher, _ = rc4.NewCipher(key[:5])
		user := make([]byte, 32)
		cipher.XORKeyStream(user, pad)
		objects = append(objects, fmt.Sprintf("<< /Filter /Standard /V 1 /R 2 /O <%x> /U <%x> /P -4 >>", owner, user))
		encryptRef = fmt.Sprintf(" /Encrypt %d 0 R", len(objects))
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	if padding > 0 {
		out.WriteByte('%')
		out.Write(bytes.Repeat([]byte{'x'}, padding))
		out.WriteByte('\n')
	}
	offsets := []int{0}
	for i, obj := range objects {
		offsets = append(offsets, out.Len())
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&out, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R /ID [<%s><%s>]%s >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), hex.EncodeToString(id), hex.EncodeToString(id), encryptRef, xref)
	return out.Bytes()
}

func sizedPDF(size int) []byte {
	// Fixed-point padding accounts for the varying number of startxref digits.
	padding := size - len(pdfFixture(1, 0, nil)) - 2
	for i := 0; i < 10; i++ {
		b := pdfFixture(1, padding, nil)
		delta := size - len(b)
		if delta == 0 {
			return b
		}
		padding += delta
	}
	panic("PDF size fixture did not converge")
}
