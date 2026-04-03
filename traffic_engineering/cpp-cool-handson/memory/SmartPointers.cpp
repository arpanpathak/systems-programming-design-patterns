// Smart pointers: unique_ptr, shared_ptr, weak_ptr
#include <iostream>
#include <memory>
#include <vector>
#include <string>

class Resource {
    std::string name_;
public:
    explicit Resource(std::string name) : name_(std::move(name)) {
        std::cout << "  [+] Resource created: " << name_ << "\n";
    }
    ~Resource() {
        std::cout << "  [-] Resource destroyed: " << name_ << "\n";
    }
    void use() const { std::cout << "  [*] Using: " << name_ << "\n"; }
    const std::string& name() const { return name_; }
};

// --- unique_ptr: exclusive ownership ---
void unique_ptr_demo() {
    std::cout << "\n=== unique_ptr (exclusive ownership) ===\n";
    auto res = std::make_unique<Resource>("UniqueFile");
    res->use();

    // Transfer ownership
    auto res2 = std::move(res);
    std::cout << "  Moved ownership. Original is "
              << (res ? "valid" : "null") << "\n";
    res2->use();
    // res2 automatically destroyed at scope exit
}

// --- shared_ptr: shared ownership ---
void shared_ptr_demo() {
    std::cout << "\n=== shared_ptr (shared ownership) ===\n";
    auto res = std::make_shared<Resource>("SharedDB");
    std::cout << "  ref_count = " << res.use_count() << "\n";

    {
        auto res2 = res;  // shared
        std::cout << "  ref_count = " << res.use_count() << "\n";
        res2->use();
    }
    std::cout << "  ref_count after scope = " << res.use_count() << "\n";
}

// --- weak_ptr: non-owning observer (breaks cycles) ---
void weak_ptr_demo() {
    std::cout << "\n=== weak_ptr (observer, breaks cycles) ===\n";
    std::weak_ptr<Resource> weak;

    {
        auto shared = std::make_shared<Resource>("WeakTarget");
        weak = shared;
        if (auto locked = weak.lock()) {
            std::cout << "  weak_ptr is valid: ";
            locked->use();
        }
    }

    if (weak.expired()) {
        std::cout << "  weak_ptr expired (resource destroyed)\n";
    }
}

// --- Custom deleter ---
void custom_deleter_demo() {
    std::cout << "\n=== Custom Deleter ===\n";
    auto deleter = [](Resource* r) {
        std::cout << "  [Custom] Cleaning up " << r->name() << "\n";
        delete r;
    };
    std::unique_ptr<Resource, decltype(deleter)> res(
        new Resource("CustomManaged"), deleter);
    res->use();
}

int main() {
    unique_ptr_demo();
    shared_ptr_demo();
    weak_ptr_demo();
    custom_deleter_demo();
    std::cout << "\n=== All resources cleaned up ===\n";
    return 0;
}
