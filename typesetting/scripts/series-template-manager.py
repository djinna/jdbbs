#!/usr/bin/env python3
"""
Series Template Manager
Manages templates and settings for book series to ensure consistency
"""

import json
import shutil
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Optional
import hashlib


class SeriesTemplateManager:
    """Manages series templates and their versioning"""
    
    def __init__(self, base_path: str = "./book-production/series-templates"):
        self.base_path = Path(base_path)
        self.base_path.mkdir(parents=True, exist_ok=True)
        self.registry_file = self.base_path / "series_registry.json"
        self.registry = self._load_registry()
        
    def _load_registry(self) -> Dict:
        """Load or create series registry"""
        if self.registry_file.exists():
            with open(self.registry_file, 'r') as f:
                return json.load(f)
        return {"series": {}, "version": "1.0"}
    
    def _save_registry(self):
        """Save registry to disk"""
        with open(self.registry_file, 'w') as f:
            json.dump(self.registry, f, indent=2)
    
    def create_series(self, series_name: str, series_id: str = None) -> str:
        """Create a new series entry"""
        if not series_id:
            # Generate ID from name
            series_id = series_name.lower().replace(' ', '-')
            series_id = ''.join(c for c in series_id if c.isalnum() or c == '-')
        
        if series_id in self.registry["series"]:
            raise ValueError(f"Series '{series_id}' already exists")
        
        series_path = self.base_path / series_id
        series_path.mkdir(exist_ok=True)
        
        self.registry["series"][series_id] = {
            "name": series_name,
            "created": datetime.now().isoformat(),
            "updated": datetime.now().isoformat(),
            "books": [],
            "current_template_version": None,
            "template_history": []
        }
        
        self._save_registry()
        return series_id
    
    def save_series_template(self, series_id: str, template_data: Dict, 
                           template_docx_path: str = None) -> str:
        """Save a new template version for a series"""
        if series_id not in self.registry["series"]:
            raise ValueError(f"Series '{series_id}' not found")
        
        series = self.registry["series"][series_id]
        series_path = self.base_path / series_id
        
        # Generate version timestamp
        version = datetime.now().strftime("%Y-%m-%d %H%M")
        
        # Save template data (book spec)
        template_file = series_path / f"template_spec {version}.json"
        with open(template_file, 'w') as f:
            json.dump(template_data, f, indent=2)
        
        # Copy Word template if provided
        if template_docx_path and Path(template_docx_path).exists():
            docx_dest = series_path / f"template {version}.docx"
            shutil.copy2(template_docx_path, docx_dest)
        
        # Update registry
        template_entry = {
            "version": version,
            "created": datetime.now().isoformat(),
            "spec_file": template_file.name,
            "docx_file": f"template {version}.docx" if template_docx_path else None,
            "checksum": self._calculate_checksum(template_data)
        }
        
        series["template_history"].append(template_entry)
        series["current_template_version"] = version
        series["updated"] = datetime.now().isoformat()
        
        self._save_registry()
        return version
    
    def get_series_template(self, series_id: str, version: str = None) -> Dict:
        """Retrieve a series template (latest by default)"""
        if series_id not in self.registry["series"]:
            raise ValueError(f"Series '{series_id}' not found")
        
        series = self.registry["series"][series_id]
        series_path = self.base_path / series_id
        
        if not version:
            version = series["current_template_version"]
        
        if not version:
            raise ValueError(f"No template found for series '{series_id}'")
        
        # Find template in history
        template_entry = None
        for entry in series["template_history"]:
            if entry["version"] == version:
                template_entry = entry
                break
        
        if not template_entry:
            raise ValueError(f"Version '{version}' not found for series '{series_id}'")
        
        # Load spec file
        spec_path = series_path / template_entry["spec_file"]
        with open(spec_path, 'r') as f:
            spec_data = json.load(f)
        
        # Add paths to files
        result = {
            "series_id": series_id,
            "series_name": series["name"],
            "version": version,
            "spec": spec_data,
            "spec_path": str(spec_path),
        }
        
        if template_entry.get("docx_file"):
            docx_path = series_path / template_entry["docx_file"]
            if docx_path.exists():
                result["docx_path"] = str(docx_path)
        
        return result
    
    def add_book_to_series(self, series_id: str, book_data: Dict) -> None:
        """Register a book as part of a series"""
        if series_id not in self.registry["series"]:
            raise ValueError(f"Series '{series_id}' not found")
        
        series = self.registry["series"][series_id]
        
        book_entry = {
            "title": book_data.get("title"),
            "author": book_data.get("author"),
            "project_id": book_data.get("project_id"),
            "volume_number": book_data.get("volume_number"),
            "added": datetime.now().isoformat(),
            "template_version_used": series["current_template_version"]
        }
        
        series["books"].append(book_entry)
        series["updated"] = datetime.now().isoformat()
        
        self._save_registry()
    
    def list_series(self) -> List[Dict]:
        """List all series with summary info"""
        result = []
        for series_id, series_data in self.registry["series"].items():
            result.append({
                "series_id": series_id,
                "name": series_data["name"],
                "book_count": len(series_data["books"]),
                "current_template_version": series_data["current_template_version"],
                "created": series_data["created"],
                "updated": series_data["updated"]
            })
        return sorted(result, key=lambda x: x["name"])
    
    def get_series_info(self, series_id: str) -> Dict:
        """Get detailed information about a series"""
        if series_id not in self.registry["series"]:
            raise ValueError(f"Series '{series_id}' not found")
        
        series = self.registry["series"][series_id]
        return {
            "series_id": series_id,
            "name": series["name"],
            "created": series["created"],
            "updated": series["updated"],
            "books": series["books"],
            "current_template_version": series["current_template_version"],
            "template_versions": [
                {
                    "version": t["version"],
                    "created": t["created"],
                    "is_current": t["version"] == series["current_template_version"]
                }
                for t in series["template_history"]
            ]
        }
    
    def compare_templates(self, series_id: str, version1: str, version2: str) -> Dict:
        """Compare two template versions"""
        template1 = self.get_series_template(series_id, version1)
        template2 = self.get_series_template(series_id, version2)
        
        # Deep comparison of specs
        differences = self._compare_dicts(template1["spec"], template2["spec"])
        
        return {
            "series_id": series_id,
            "version1": version1,
            "version2": version2,
            "differences": differences
        }
    
    def _calculate_checksum(self, data: Dict) -> str:
        """Calculate checksum for template data"""
        json_str = json.dumps(data, sort_keys=True)
        return hashlib.sha256(json_str.encode()).hexdigest()[:16]
    
    def _compare_dicts(self, d1: Dict, d2: Dict, path: str = "") -> List[Dict]:
        """Deep comparison of two dictionaries"""
        differences = []
        
        all_keys = set(d1.keys()) | set(d2.keys())
        
        for key in all_keys:
            current_path = f"{path}.{key}" if path else key
            
            if key not in d1:
                differences.append({
                    "path": current_path,
                    "type": "added",
                    "value": d2[key]
                })
            elif key not in d2:
                differences.append({
                    "path": current_path,
                    "type": "removed",
                    "value": d1[key]
                })
            elif d1[key] != d2[key]:
                if isinstance(d1[key], dict) and isinstance(d2[key], dict):
                    # Recursive comparison
                    differences.extend(self._compare_dicts(d1[key], d2[key], current_path))
                else:
                    differences.append({
                        "path": current_path,
                        "type": "changed",
                        "old_value": d1[key],
                        "new_value": d2[key]
                    })
        
        return differences


class SeriesTemplateAPI:
    """High-level API for series template management"""
    
    def __init__(self, manager: SeriesTemplateManager):
        self.manager = manager
    
    def create_series_from_book(self, series_name: str, book_spec: Dict,
                               template_docx: str = None) -> Dict:
        """Create a new series using a book as the template"""
        # Create series
        series_id = self.manager.create_series(series_name)
        
        # Save the book's spec as the series template
        version = self.manager.save_series_template(
            series_id, book_spec, template_docx
        )
        
        # Add the book to the series
        self.manager.add_book_to_series(series_id, {
            "title": book_spec.get("metadata", {}).get("title"),
            "author": book_spec.get("metadata", {}).get("author"),
            "volume_number": 1
        })
        
        return {
            "series_id": series_id,
            "template_version": version,
            "message": f"Series '{series_name}' created with template from first book"
        }
    
    def apply_series_template_to_book(self, series_id: str, 
                                    book_transmittal: Dict) -> Dict:
        """Apply series template to a new book in the series"""
        # Get the series template
        template = self.manager.get_series_template(series_id)
        
        # Merge template with book-specific data
        book_spec = template["spec"].copy()
        
        # Update metadata with book-specific info
        book_spec["metadata"].update({
            "title": book_transmittal.get("title"),
            "subtitle": book_transmittal.get("subtitle"),
            "author": book_transmittal.get("author"),
            "isbn_paper": book_transmittal.get("isbn_paper"),
            "isbn_cloth": book_transmittal.get("isbn_cloth"),
        })
        
        # Add to series
        self.manager.add_book_to_series(series_id, {
            "title": book_transmittal.get("title"),
            "author": book_transmittal.get("author"),
            "project_id": book_transmittal.get("project_id"),
            "volume_number": book_transmittal.get("volume_number")
        })
        
        return {
            "book_spec": book_spec,
            "series_id": series_id,
            "template_version": template["version"],
            "template_docx": template.get("docx_path")
        }
    
    def update_series_template(self, series_id: str, updated_spec: Dict,
                             updated_docx: str = None, reason: str = None) -> Dict:
        """Update a series template with a new version"""
        # Get current template for comparison
        current = self.manager.get_series_template(series_id)
        
        # Save new version
        new_version = self.manager.save_series_template(
            series_id, updated_spec, updated_docx
        )
        
        # Compare changes
        differences = self.manager.compare_templates(
            series_id, current["version"], new_version
        )
        
        return {
            "series_id": series_id,
            "old_version": current["version"],
            "new_version": new_version,
            "changes": len(differences["differences"]),
            "reason": reason
        }


# CLI interface
def main():
    import argparse
    
    parser = argparse.ArgumentParser(description='Series Template Manager')
    subparsers = parser.add_subparsers(dest='command', help='Commands')
    
    # List series
    list_parser = subparsers.add_parser('list', help='List all series')
    
    # Create series
    create_parser = subparsers.add_parser('create', help='Create new series')
    create_parser.add_argument('name', help='Series name')
    create_parser.add_argument('--id', help='Series ID (auto-generated if not provided)')
    
    # Show series info
    info_parser = subparsers.add_parser('info', help='Show series information')
    info_parser.add_argument('series_id', help='Series ID')
    
    # Get template
    get_parser = subparsers.add_parser('get-template', help='Get series template')
    get_parser.add_argument('series_id', help='Series ID')
    get_parser.add_argument('--version', help='Template version (latest by default)')
    get_parser.add_argument('--output', help='Output file for spec JSON')
    
    args = parser.parse_args()
    
    manager = SeriesTemplateManager()
    
    if args.command == 'list':
        series_list = manager.list_series()
        if not series_list:
            print("No series found")
        else:
            print(f"{'Series ID':<20} {'Name':<30} {'Books':<6} {'Updated'}")
            print("-" * 70)
            for s in series_list:
                print(f"{s['series_id']:<20} {s['name']:<30} "
                      f"{s['book_count']:<6} {s['updated'][:10]}")
    
    elif args.command == 'create':
        series_id = manager.create_series(args.name, args.id)
        print(f"Created series '{args.name}' with ID: {series_id}")
    
    elif args.command == 'info':
        info = manager.get_series_info(args.series_id)
        print(f"Series: {info['name']} (ID: {info['series_id']})")
        print(f"Created: {info['created'][:10]}")
        print(f"Books: {len(info['books'])}")
        print(f"Current template: {info['current_template_version'] or 'None'}")
        
        if info['books']:
            print("\nBooks in series:")
            for i, book in enumerate(info['books'], 1):
                print(f"  {i}. {book['title']} by {book['author']}")
    
    elif args.command == 'get-template':
        template = manager.get_series_template(args.series_id, args.version)
        
        if args.output:
            with open(args.output, 'w') as f:
                json.dump(template['spec'], f, indent=2)
            print(f"Template saved to {args.output}")
        else:
            print(json.dumps(template['spec'], indent=2))
        
        if template.get('docx_path'):
            print(f"\nWord template available at: {template['docx_path']}")


if __name__ == '__main__':
    main()