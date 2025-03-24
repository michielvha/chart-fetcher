[GitVersion](https://gitversion.net/docs/reference/configuration) provides several **modes** that determine how version numbers are calculated based on your branching and commit strategy. These modes cater to different workflows, like feature branching, continuous integration, or release management. Below are all the available modes:

---

### **1. Continuous Delivery**
```yaml
mode: ContinuousDelivery
```

#### **Description:**
- This mode is designed for workflows where you want **prerelease versions** (e.g., `1.0.0-beta.1`) to be generated consistently until the changes are merged into the main branch.
- It keeps the version increment predictable but ensures it’s clearly marked as prerelease.

#### **Use Case:**
- CI pipelines where you frequently release snapshots or artifacts for testing before a stable release.
- Example: `1.2.0-beta.1`, `1.2.0-beta.2`, ...

---

### **2. Continuous Deployment**
```yaml
mode: ContinuousDeployment
```

#### **Description:**
- This mode ensures that every commit produces a **stable version**, even on non-release branches.
- Versions are incremented automatically and don’t include prerelease tags.

#### **Use Case:**
- Environments where every commit should result in a production-ready version (e.g., microservices or continuous deployment pipelines).
- Example: `1.2.0`, `1.2.1`, `1.3.0`, ...

---

### **3. Mainline**
```yaml
mode: Mainline
```

#### **Description:**
- Suitable for **trunk-based development** workflows.
- Merges into the main branch determine version increments (e.g., major, minor, patch).
- Branches other than `main` (e.g., feature branches) do not increment the version unless merged.
- The calculated version reflects the main branch's linear progression.

#### **Use Case:**
- Teams practicing trunk-based development or working with short-lived branches.
- Example: `1.2.0`, `1.3.0` after merging a feature.

---

### **4. GitFlow (Default)**
```yaml
mode: GitFlow
```

#### **Description:**
- Follows the **GitFlow branching strategy**.
- Release branches (`release/x.x`) define the next stable version.
- `develop` produces prerelease versions (e.g., `1.2.0-alpha.1`).
- Feature branches generate unique prerelease versions (e.g., `1.2.0-feature.1`).

#### **Use Case:**
- Classic GitFlow workflows with `develop`, `release`, and `hotfix` branches.
- Example: `1.2.0-alpha.1` on `develop`, `1.2.0-beta.1` on `release`.

---

### **5. Continuous Integration**
```yaml
mode: ContinuousIntegration
```

#### **Description:**
- Similar to **Continuous Delivery**, but every build generates **incremental versions** that are not tied to a stable or prerelease strategy.
- Versions are based on the build metadata (e.g., `+001`).

#### **Use Case:**
- CI environments where the main focus is distinguishing builds without strict adherence to a branching model.
- Example: `1.2.0+001`, `1.2.0+002`.

---

### **6. None**
```yaml
mode: None
```

#### **Description:**
- Disables version calculation by GitVersion.
- Useful if you want to rely entirely on custom scripts or external tools for versioning.

#### **Use Case:**
- Projects that don’t need GitVersion’s automatic calculation but may still require some parts of the toolchain.

---

### **Mode Comparison Table**

| **Mode**                  | **Stable Versions** | **Prerelease Versions** | **Branching Strategy** | **Best For**                       |
|---------------------------|---------------------|--------------------------|-------------------------|-------------------------------------|
| **Continuous Delivery**   | No                 | Yes                     | All                    | Snapshots for testing.             |
| **Continuous Deployment** | Yes                | No                      | All                    | Production-ready builds.           |
| **Mainline**              | Yes                | No                      | Trunk-based            | Trunk-based development.           |
| **GitFlow**               | Yes                | Yes                     | GitFlow                | Traditional GitFlow branching.     |
| **Continuous Integration**| Yes                | Optional                | Flexible               | Incremental build metadata.        |
| **None**                  | No                 | No                      | Custom                 | Manual or custom versioning.       |

---

### **How to Choose the Right Mode**
- **Continuous Delivery**: Use if you’re releasing prerelease versions frequently but don’t want them marked as stable.
- **Continuous Deployment**: Use if every commit should produce a stable, production-ready version.
- **Mainline**: Use if you follow trunk-based development and want simple, linear versioning.
- **GitFlow**: Use if your workflow includes `develop`, `release`, and `hotfix` branches.
- **Continuous Integration**: Use if you don’t need complex versioning but want unique identifiers for builds.
- **None**: Use if you don’t want GitVersion to handle versioning at all.