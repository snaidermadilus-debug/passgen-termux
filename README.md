# 🔐 Passgen Termux

Gerador de senhas seguras em **Go** para **Termux**. Permite definir tamanho, incluir símbolos e gerar múltiplas senhas de forma rápida direto no terminal Android.

---

## 🌟 Funcionalidades

- ✅ Gera senhas seguras e aleatórias  
- ✅ Define o **tamanho da senha**  
- ✅ Incluir símbolos especiais (opcional)  
- ✅ Gerar múltiplas senhas de uma vez  
- ✅ Funciona direto no **Termux**  
./passgen -l 32 -s -count 5
               👈
           você pode modificar esse código também você pode botar o número que você quer esse código aqui
---

## 📥 Como baixar e usar (Linha Única)

Para instalar e rodar direto no Termux, cole a linha abaixo:

```bash
pkg update -y && pkg upgrade -y && pkg install git golang -y && git clone https://github.com/snaidermadilus-debug/passgen-termux.git && cd passgen-termux && go build -o passgen main.go && ./passgen -l 24 -s -count 3





./passgen -l 32 -s -count 5
               