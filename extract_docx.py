#!/usr/bin/env python3
import sys
import re
from docx import Document

def extract_appendix_e(doc_path):
    """提取文档中的附录E内容"""
    try:
        doc = Document(doc_path)
        content = []
        in_appendix_e = False
        appendix_e_title = None
        
        for para in doc.paragraphs:
            text = para.text.strip()
            
            # 查找附录E
            if re.search(r'附录\s*[Ee]|Appendix\s*[Ee]', text):
                in_appendix_e = True
                appendix_e_title = text
                content.append(f"# {text}")
                continue
                
            # 如果已经在附录E中，收集内容
            if in_appendix_e:
                # 检查是否到了下一个附录或章节
                if re.search(r'附录\s*[Ff]|Appendix\s*[Ff]', text):
                    break
                    
                if text:
                    content.append(text)
        
        return appendix_e_title, "\n".join(content)
        
    except Exception as e:
        return None, f"提取文档时出错: {str(e)}"

def main():
    doc_path = "/app/logsift/docs/环境 & prompt & traj 标注标准.docx"
    
    print("正在提取文档内容...")
    title, content = extract_appendix_e(doc_path)
    
    if title:
        print(f"\n找到: {title}")
        print("\n" + "="*80)
        print(content)
        print("="*80)
        
        # 保存到文件
        with open("/app/logsift/.trae/appendix_e_content.txt", "w", encoding="utf-8") as f:
            f.write(content)
        print(f"\n内容已保存到: /app/logsift/.trae/appendix_e_content.txt")
    else:
        print("未找到附录E内容")
        if content:
            print(f"错误信息: {content}")

if __name__ == "__main__":
    main()